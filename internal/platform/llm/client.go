package llm

import (
	"context"
	"fmt"
	"strings"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"

	"github.com/ZyqAlwaysCool/anvil/internal/platform/config"
)

type Client struct {
	cfg    config.LLMConfig
	client *openai.Client
}

// New 仅在 LLM_ENABLED=true 时装配客户端；关闭时不阻塞 Server/Worker 启动。
func New(cfg config.LLMConfig) (*Client, error) {
	if !cfg.Enabled {
		return nil, nil
	}

	opts := []option.RequestOption{option.WithAPIKey(cfg.APIKey)}
	if strings.TrimSpace(cfg.BaseURL) != "" {
		opts = append(opts, option.WithBaseURL(cfg.BaseURL))
	}
	client := openai.NewClient(opts...)
	return &Client{cfg: cfg, client: &client}, nil
}

func (c *Client) Generate(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error) {
	if c == nil || c.client == nil {
		return nil, fmt.Errorf("llm client is not configured")
	}

	params, err := BuildCompletionParams(req, c.cfg)
	if err != nil {
		return nil, err
	}

	callCtx := ctx
	if c.cfg.Timeout > 0 {
		timeoutCtx, cancel := context.WithTimeout(ctx, c.cfg.Timeout)
		defer cancel()
		callCtx = timeoutCtx
	}

	completion, err := c.client.Chat.Completions.New(callCtx, params)
	if err != nil {
		return nil, fmt.Errorf("llm generate failed: %w", err)
	}

	return mapCompletion(completion)
}

func (c *Client) GenerateStream(ctx context.Context, req *GenerateRequest) (<-chan StreamChunk, <-chan error) {
	chunkCh := make(chan StreamChunk)
	errCh := make(chan error, 1)

	go func() {
		defer close(chunkCh)
		defer close(errCh)

		if c == nil || c.client == nil {
			errCh <- fmt.Errorf("llm client is not configured")
			return
		}

		params, err := BuildCompletionParams(req, c.cfg)
		if err != nil {
			errCh <- err
			return
		}

		callCtx := ctx
		if c.cfg.Timeout > 0 {
			timeoutCtx, cancel := context.WithTimeout(ctx, c.cfg.Timeout)
			defer cancel()
			callCtx = timeoutCtx
		}

		// 使用 openai-go 真实 streaming API，按增量下发 Delta，结束时发送 Done=true。
		stream := c.client.Chat.Completions.NewStreaming(callCtx, params)
		defer stream.Close()

		acc := openai.ChatCompletionAccumulator{}
		for stream.Next() {
			if err := callCtx.Err(); err != nil {
				errCh <- err
				return
			}

			chunk := stream.Current()
			acc.AddChunk(chunk)
			if len(chunk.Choices) == 0 {
				continue
			}

			delta := chunk.Choices[0].Delta
			if delta.Refusal != "" {
				errCh <- &RefusalError{Response: &GenerateResponse{Refusal: delta.Refusal}}
				return
			}
			if delta.Content != "" {
				chunkCh <- StreamChunk{Delta: delta.Content}
			}
		}

		if err := stream.Err(); err != nil {
			errCh <- fmt.Errorf("llm stream failed: %w", err)
			return
		}

		completion := acc.ChatCompletion
		resp, err := mapCompletion(&completion)
		if err != nil {
			errCh <- err
			return
		}
		chunkCh <- StreamChunk{Done: true, Usage: resp.Usage, Refusal: resp.Refusal}
	}()

	return chunkCh, errCh
}

func mapCompletion(completion *openai.ChatCompletion) (*GenerateResponse, error) {
	if completion == nil || len(completion.Choices) == 0 {
		return nil, ErrEmptyChoices
	}

	choice := completion.Choices[0]
	if choice.Message.Refusal != "" {
		return &GenerateResponse{Refusal: choice.Message.Refusal}, &RefusalError{
			Response: &GenerateResponse{Refusal: choice.Message.Refusal},
		}
	}

	toolCalls := make([]ToolCall, 0, len(choice.Message.ToolCalls))
	for _, call := range choice.Message.ToolCalls {
		if strings.TrimSpace(call.ID) == "" || strings.TrimSpace(call.Function.Name) == "" {
			return nil, fmt.Errorf("%w: missing tool call id or name", ErrInvalidToolCall)
		}
		toolCalls = append(toolCalls, ToolCall{
			ID:        call.ID,
			Name:      call.Function.Name,
			Arguments: call.Function.Arguments,
		})
	}

	var usage *Usage
	if completion.Usage.PromptTokens > 0 || completion.Usage.CompletionTokens > 0 {
		usage = &Usage{
			InputTokens:  int(completion.Usage.PromptTokens),
			OutputTokens: int(completion.Usage.CompletionTokens),
			TotalTokens:  int(completion.Usage.TotalTokens),
		}
	}

	return &GenerateResponse{
		Content:   choice.Message.Content,
		ToolCalls: toolCalls,
		Usage:     usage,
	}, nil
}
