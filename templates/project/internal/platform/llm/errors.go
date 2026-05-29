//go:build ignore

package llm

import "errors"

var (
	// ErrRefusal 表示模型明确拒答；调用方应读取 GenerateResponse.Refusal，而不是把它当作普通内部错误。
	ErrRefusal = errors.New("llm refusal")
	// ErrEmptyChoices 表示底层返回空 choices。
	ErrEmptyChoices = errors.New("llm empty choices")
	// ErrInvalidToolCall 表示工具调用结果不完整。
	ErrInvalidToolCall = errors.New("llm invalid tool call")
)

// RefusalError 保留拒答语义：Response 中 Refusal 字段已填充，Err 为 ErrRefusal。
type RefusalError struct {
	Response *GenerateResponse
}

func (e *RefusalError) Error() string {
	if e == nil || e.Response == nil {
		return ErrRefusal.Error()
	}
	return "llm refusal: " + e.Response.Refusal
}

func (e *RefusalError) Unwrap() error {
	return ErrRefusal
}
