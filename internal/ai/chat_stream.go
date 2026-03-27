package ai

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"go.jetify.com/ai"
	"go.jetify.com/ai/api"
)

// Message is the public type callers use to build conversation history.
type Message struct {
	Role    string // "system" | "user" | "assistant"
	Content string
}

// StreamResult carries the output of a single stream request.
type StreamResult struct {
	// FullText accumulates all TextDelta events.
	FullText string
	// InputTokens and OutputTokens are populated from the FinishEvent.
	InputTokens  int
	OutputTokens int
}

// ChatStream sends messages to the model and calls onChunk for every text delta.
// It returns StreamResult after the stream closes, or an error.
//
// Design notes:
//   - model is injected; no package-level state.
//   - ctx should already carry a deadline (e.g. 10-second timeout set by caller).
//   - onChunk must be cheap (e.g. fmt.Print) because it runs on the hot path.
func ChatStream(
	ctx context.Context,
	model api.LanguageModel,
	messages []Message,
	onChunk func(text string),
) (*StreamResult, error) {
	prompt, err := buildPrompt(messages)
	if err != nil {
		return nil, err
	}

	resp, err := ai.StreamText(ctx, prompt, ai.WithModel(model))
	if err != nil {
		return nil, wrapAPIError(err)
	}

	var result StreamResult
	for event := range resp.Stream {
		switch e := event.(type) {
		case *api.TextDeltaEvent:
			result.FullText += e.TextDelta
			if onChunk != nil {
				onChunk(e.TextDelta)
			}
		case *api.FinishEvent:
			result.InputTokens = e.Usage.InputTokens
			result.OutputTokens = e.Usage.OutputTokens
		case *api.ErrorEvent:
			// Errors mid-stream arrive as events, not as Go errors.
			return nil, fmt.Errorf("stream error: %v", e.Err)
		}
	}

	return &result, nil
}

// buildPrompt converts the caller's []Message into the api.Message slice the SDK expects.
func buildPrompt(messages []Message) ([]api.Message, error) {
	out := make([]api.Message, 0, len(messages))
	for _, m := range messages {
		switch m.Role {
		case "system":
			out = append(out, &api.SystemMessage{Content: m.Content})
		case "user":
			out = append(out, &api.UserMessage{
				Content: []api.ContentBlock{&api.TextBlock{Text: m.Content}},
			})
		case "assistant":
			out = append(out, &api.AssistantMessage{
				Content: []api.ContentBlock{&api.TextBlock{Text: m.Content}},
			})
		default:
			return nil, fmt.Errorf("unknown message role %q", m.Role)
		}
	}
	return out, nil
}

// wrapAPIError enriches API errors with actionable context.
// Use errors.Is / errors.As in call-sites to branch on specific subtypes.
func wrapAPIError(err error) error {
	var apiErr *api.APICallError
	if errors.As(err, &apiErr) {
		switch apiErr.StatusCode {
		case http.StatusTooManyRequests: // 429
			return fmt.Errorf("rate limit exceeded (429) — back off and retry: %w", err)
		case http.StatusUnauthorized: // 401
			return fmt.Errorf("invalid API key (401) — check OPENAI_API_KEY: %w", err)
		case http.StatusUnprocessableEntity: // 422
			return fmt.Errorf("model rejected the request (422): %w", err)
		}
		if apiErr.IsRetryable() {
			return fmt.Errorf("transient API error (%d), safe to retry: %w", apiErr.StatusCode, err)
		}
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("request timed out: %w", err)
	}
	if errors.Is(err, io.EOF) {
		return fmt.Errorf("stream closed unexpectedly: %w", err)
	}

	return err
}
