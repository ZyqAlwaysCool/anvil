package test

import (
	"testing"

	"github.com/openai/openai-go/packages/param"
	"github.com/zyq/anvil/internal/platform/config"
	"github.com/zyq/anvil/internal/platform/llm"
)

func TestBuildCompletionParamsToolsAndToolChoice(t *testing.T) {
	params, err := llm.BuildCompletionParams(&llm.GenerateRequest{
		Model: "gpt-4o-mini",
		Tools: []llm.ToolSpec{{
			Name:        "lookup",
			Description: "lookup data",
			InputSchema: map[string]any{"type": "object"},
		}},
		ToolChoice: "required",
	}, config.LLMConfig{Model: "gpt-4o-mini"})
	if err != nil {
		t.Fatalf("build params: %v", err)
	}
	if len(params.Tools) != 1 {
		t.Fatalf("expected tools mapped, got %d", len(params.Tools))
	}
	if param.IsOmitted(params.ToolChoice.OfAuto) {
		t.Fatal("expected required tool choice mapped to OfAuto")
	}
}

func TestBuildCompletionParamsResponseFormatJSONSchema(t *testing.T) {
	params, err := llm.BuildCompletionParams(&llm.GenerateRequest{
		Model: "gpt-4o-mini",
		ResponseFormat: &llm.JSONSchemaSpec{
			Name:   "greeting",
			Strict: true,
			Schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"greeting": map[string]any{"type": "string"},
				},
			},
		},
	}, config.LLMConfig{Model: "gpt-4o-mini"})
	if err != nil {
		t.Fatalf("build params: %v", err)
	}
	if params.ResponseFormat.OfJSONSchema == nil {
		t.Fatal("expected json schema response format mapped")
	}
	if params.ResponseFormat.OfJSONSchema.JSONSchema.Name != "greeting" {
		t.Fatalf("unexpected schema name: %s", params.ResponseFormat.OfJSONSchema.JSONSchema.Name)
	}
}

func TestBuildCompletionParamsNamedToolChoice(t *testing.T) {
	params, err := llm.BuildCompletionParams(&llm.GenerateRequest{
		Model: "gpt-4o-mini",
		Tools: []llm.ToolSpec{{
			Name:        "lookup",
			Description: "lookup data",
			InputSchema: map[string]any{"type": "object"},
		}},
		ToolChoice: "lookup",
	}, config.LLMConfig{Model: "gpt-4o-mini"})
	if err != nil {
		t.Fatalf("build params: %v", err)
	}
	if params.ToolChoice.OfChatCompletionNamedToolChoice == nil {
		t.Fatal("expected named function tool choice")
	}
	if params.ToolChoice.OfChatCompletionNamedToolChoice.Function.Name != "lookup" {
		t.Fatalf("unexpected function name: %s", params.ToolChoice.OfChatCompletionNamedToolChoice.Function.Name)
	}
}

func TestBuildCompletionParamsToolChoiceNone(t *testing.T) {
	params, err := llm.BuildCompletionParams(&llm.GenerateRequest{
		Model:      "gpt-4o-mini",
		ToolChoice: "none",
	}, config.LLMConfig{Model: "gpt-4o-mini"})
	if err != nil {
		t.Fatalf("build params: %v", err)
	}
	if param.IsOmitted(params.ToolChoice.OfAuto) {
		t.Fatal("expected none tool choice mapped to OfAuto")
	}
}
