package ai

import (
	"errors"
	"fmt"
	"os"

	"go.jetify.com/ai/api"
	"go.jetify.com/ai/provider/openai"
)

// KnownModel describes a selectable LLM.
type KnownModel struct {
	ID       string
	Provider string
	Label    string
}

// AvailableModels is the catalogue of API-based models (requires an API key).
// Claude models are NOT listed here — they route through the Claude CLI instead,
// which uses the user's existing CLI auth and needs no separate ANTHROPIC_API_KEY.
// See internal/claude.ListCLIModels() for the Claude model catalogue.
var AvailableModels = []KnownModel{
	// OpenAI — requires OPENAI_API_KEY
	{ID: "gpt-4o", Provider: "openai", Label: "OpenAI GPT-4o"},
	{ID: "gpt-4o-mini", Provider: "openai", Label: "OpenAI GPT-4o mini"},
	{ID: "gpt-3.5-turbo", Provider: "openai", Label: "OpenAI GPT-3.5 Turbo"},
	{ID: "o3-mini", Provider: "openai", Label: "OpenAI o3-mini"},
}

// NewModel builds an api.LanguageModel for the given modelID.
// The provider is inferred from AvailableModels; the required env var is validated early.
func NewModel(modelID string) (api.LanguageModel, error) {
	var entry KnownModel
	for _, m := range AvailableModels {
		if m.ID == modelID {
			entry = m
			break
		}
	}
	if entry.ID == "" {
		return nil, fmt.Errorf("unknown model %q: run with --list-models to see options", modelID)
	}

	switch entry.Provider {
	case "openai":
		if os.Getenv("OPENAI_API_KEY") == "" {
			return nil, errors.New("OPENAI_API_KEY is not set")
		}
		// openai.NewClient() picks up OPENAI_API_KEY automatically.
		return openai.NewLanguageModel(modelID), nil

	default:
		return nil, fmt.Errorf("no provider registered for %q", entry.Provider)
	}
}

// ListModels returns the catalogue of models available in this build.
func ListModels() []KnownModel {
	return AvailableModels
}
