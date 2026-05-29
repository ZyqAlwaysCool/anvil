//go:build ignore

package example

type CreateTaskRequest struct {
	Name string `json:"name"`
}

type GreetResult struct {
	Greeting string `json:"greeting"`
}

type CreateTaskResponse struct {
	TaskID string `json:"task_id"`
}

type QueryTaskResponse struct {
	TaskID string       `json:"task_id"`
	Status string       `json:"status"`
	Result *GreetResult `json:"result,omitempty"`
}
