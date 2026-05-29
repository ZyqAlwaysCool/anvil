package llm

type Message struct {
	Role    string
	Content string
}

type ToolSpec struct {
	Name        string
	Description string
	InputSchema map[string]any
}

type ToolCall struct {
	ID        string
	Name      string
	Arguments string
}

type JSONSchemaSpec struct {
	Name   string
	Schema map[string]any
	Strict bool
}

type Usage struct {
	InputTokens  int
	OutputTokens int
	TotalTokens  int
}

type GenerateRequest struct {
	Model          string
	System         string
	Messages       []Message
	Temperature    *float64
	MaxTokens      *int
	Tools          []ToolSpec
	ToolChoice     string
	ResponseFormat *JSONSchemaSpec
}

type GenerateResponse struct {
	Content   string
	ToolCalls []ToolCall
	Usage     *Usage
	Refusal   string
}

type StreamChunk struct {
	Delta   string
	Done    bool
	Usage   *Usage
	Refusal string
}
