package output

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/invopop/jsonschema"
)

// Schema 从 Go 类型生成 JSON Schema；失败必须返回错误，供 LLM 结构化输出约束使用。
func Schema[T any]() (map[string]any, error) {
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}
	schema := reflector.Reflect(new(T))
	raw, err := json.Marshal(schema)
	if err != nil {
		return nil, fmt.Errorf("marshal json schema: %w", err)
	}
	var result map[string]any
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("unmarshal json schema: %w", err)
	}
	return result, nil
}

// Parse 只接受 JSON 对象/数组；失败时返回摘要，避免把整段模型原文写入错误链。
func Parse[T any](content string) (*T, error) {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return nil, fmt.Errorf("parse structured output failed: empty content")
	}
	if !strings.HasPrefix(trimmed, "{") && !strings.HasPrefix(trimmed, "[") {
		summary := summarize(trimmed)
		return nil, fmt.Errorf("parse structured output failed: content is not json, summary=%s", summary)
	}

	var value T
	if err := json.Unmarshal([]byte(trimmed), &value); err != nil {
		summary := summarize(trimmed)
		return nil, fmt.Errorf("parse structured output failed: %v, summary=%s", err, summary)
	}
	return &value, nil
}

func summarize(content string) string {
	const maxLen = 120
	if len(content) <= maxLen {
		return content
	}
	return content[:maxLen] + "..."
}
