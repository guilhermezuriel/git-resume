package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	internalai "github.com/guilhermezuriel/git-resume/internal/ai"
	"go.jetify.com/ai/api"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// --- list available models -----------------------------------------------
	fmt.Println("Available models:")
	for _, m := range internalai.ListModels() {
		fmt.Printf("  %-12s  %s\n", m.ID, m.Label)
	}
	fmt.Println()

	// --- build client ---------------------------------------------------------
	model, err := internalai.NewModel("gpt-4o")
	if err != nil {
		return fmt.Errorf("init model: %w", err)
	}

	// --- conversation history -------------------------------------------------
	// The history slice grows turn-by-turn to maintain context.
	history := []internalai.Message{
		{Role: "system", Content: "You are a concise Go tutor. Keep every answer under 3 sentences."},
		{Role: "user", Content: "What is a goroutine?"},
	}

	if err := sendAndPrint(model, history, "goroutine"); err != nil {
		return err
	}

	// append the assistant reply as a placeholder so the model sees continuity
	history = append(history,
		internalai.Message{Role: "assistant", Content: "(previous answer)"},
		internalai.Message{Role: "user", Content: "How does a channel differ from a mutex?"},
	)

	if err := sendAndPrint(model, history, "channel vs mutex"); err != nil {
		return err
	}

	history = append(history,
		internalai.Message{Role: "assistant", Content: "(previous answer)"},
		internalai.Message{Role: "user", Content: "Give me a one-liner that spawns a goroutine safely."},
	)

	if err := sendAndPrint(model, history, "one-liner"); err != nil {
		return err
	}

	return nil
}

// sendAndPrint wraps ChatStream with a 10-second timeout and pretty output.
func sendAndPrint(model api.LanguageModel, messages []internalai.Message, label string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	fmt.Printf("=== %s ===\n", label)

	result, err := internalai.ChatStream(ctx, model, messages, func(chunk string) {
		fmt.Print(chunk) // stream tokens to stdout as they arrive
	})
	if err != nil {
		return handleStreamError(err)
	}

	fmt.Printf("\n[tokens: in=%d out=%d]\n\n", result.InputTokens, result.OutputTokens)
	return nil
}

// handleStreamError provides structured feedback for common failure modes.
func handleStreamError(err error) error {
	var apiErr *api.APICallError
	if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusTooManyRequests {
		// 429: advise caller to wait before retrying
		return fmt.Errorf("quota exceeded — wait and retry: %w", err)
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("10-second timeout hit — consider a longer deadline: %w", err)
	}
	return err
}
