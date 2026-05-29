package llm

import (
	"fmt"
	"strings"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/shared"

	"github.com/zyq/anvil/internal/platform/config"
)

// BuildCompletionParams 把平台契约映射到底层 openai-go 请求；边界：平台类型在此收敛，业务不得直接依赖 SDK 类型。
func BuildCompletionParams(req *GenerateRequest, cfg config.LLMConfig) (openai.ChatCompletionNewParams, error) {
	if req == nil {
		return openai.ChatCompletionNewParams{}, fmt.Errorf("generate request is nil")
	}

	model := strings.TrimSpace(req.Model)
	if model == "" {
		model = cfg.Model
	}
	if model == "" {
		return openai.ChatCompletionNewParams{}, fmt.Errorf("model is required")
	}

	messages := buildMessages(req)

	params := openai.ChatCompletionNewParams{
		Model:    openai.ChatModel(model),
		Messages: messages,
	}
	if req.Temperature != nil {
		params.Temperature = openai.Float(*req.Temperature)
	} else {
		params.Temperature = openai.Float(cfg.Temperature)
	}
	if req.MaxTokens != nil {
		params.MaxTokens = openai.Int(int64(*req.MaxTokens))
	} else if cfg.MaxTokens > 0 {
		params.MaxTokens = openai.Int(int64(cfg.MaxTokens))
	}

	toolParams, err := buildTools(req.Tools)
	if err != nil {
		return openai.ChatCompletionNewParams{}, err
	}
	if len(toolParams) > 0 {
		params.Tools = toolParams
	}

	if err := applyToolChoice(&params, req.ToolChoice); err != nil {
		return openai.ChatCompletionNewParams{}, err
	}

	if req.ResponseFormat != nil {
		if strings.TrimSpace(req.ResponseFormat.Name) == "" {
			return openai.ChatCompletionNewParams{}, fmt.Errorf("response format name is required")
		}
		if len(req.ResponseFormat.Schema) == 0 {
			return openai.ChatCompletionNewParams{}, fmt.Errorf("response format schema is required")
		}
		// JSON Schema 约束必须在请求阶段下发，不能只在响应后做 json.Valid 检查。
		params.ResponseFormat = openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONSchema: &shared.ResponseFormatJSONSchemaParam{
				JSONSchema: shared.ResponseFormatJSONSchemaJSONSchemaParam{
					Name:   req.ResponseFormat.Name,
					Schema: req.ResponseFormat.Schema,
					Strict: openai.Bool(req.ResponseFormat.Strict),
				},
			},
		}
	}

	return params, nil
}

func buildMessages(req *GenerateRequest) []openai.ChatCompletionMessageParamUnion {
	messages := make([]openai.ChatCompletionMessageParamUnion, 0, len(req.Messages)+1)
	if strings.TrimSpace(req.System) != "" {
		messages = append(messages, openai.SystemMessage(req.System))
	}
	for _, item := range req.Messages {
		switch strings.ToLower(item.Role) {
		case "assistant":
			messages = append(messages, openai.AssistantMessage(item.Content))
		case "system":
			messages = append(messages, openai.SystemMessage(item.Content))
		default:
			messages = append(messages, openai.UserMessage(item.Content))
		}
	}
	return messages
}

func buildTools(tools []ToolSpec) ([]openai.ChatCompletionToolParam, error) {
	if len(tools) == 0 {
		return nil, nil
	}

	toolParams := make([]openai.ChatCompletionToolParam, 0, len(tools))
	for _, tool := range tools {
		if strings.TrimSpace(tool.Name) == "" {
			return nil, fmt.Errorf("tool name is required")
		}
		toolParams = append(toolParams, openai.ChatCompletionToolParam{
			Function: shared.FunctionDefinitionParam{
				Name:        tool.Name,
				Description: openai.String(tool.Description),
				Parameters:  shared.FunctionParameters(tool.InputSchema),
				Strict:      openai.Bool(false),
			},
		})
	}
	return toolParams, nil
}

func applyToolChoice(params *openai.ChatCompletionNewParams, toolChoice string) error {
	choice := strings.TrimSpace(strings.ToLower(toolChoice))
	switch choice {
	case "", "auto":
		params.ToolChoice = openai.ChatCompletionToolChoiceOptionUnionParam{
			OfAuto: openai.String(string(openai.ChatCompletionToolChoiceOptionAutoAuto)),
		}
	case "none":
		params.ToolChoice = openai.ChatCompletionToolChoiceOptionUnionParam{
			OfAuto: openai.String(string(openai.ChatCompletionToolChoiceOptionAutoNone)),
		}
	case "required":
		params.ToolChoice = openai.ChatCompletionToolChoiceOptionUnionParam{
			OfAuto: openai.String(string(openai.ChatCompletionToolChoiceOptionAutoRequired)),
		}
	default:
		if len(params.Tools) == 0 {
			return fmt.Errorf("tool choice %q requires tools", toolChoice)
		}
		params.ToolChoice = openai.ChatCompletionToolChoiceOptionParamOfChatCompletionNamedToolChoice(
			openai.ChatCompletionNamedToolChoiceFunctionParam{Name: toolChoice},
		)
	}
	return nil
}
