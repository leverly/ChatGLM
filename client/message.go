package client

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type InvokeRequest struct {
	Model       string    `json:"model"`
	Prompt      []Message `json:"prompt"`
	Top_p       float32   `json:"top_p"`
	Temperature float32   `json:"temperature"`
}

type InvokeResponse struct {
	Code    int           `json:"code"`
	Msg     string        `json:"msg"`
	Success bool          `json:"success"`
	Data    *ResponseData `json:"data,omitempty"`
}

type ResponseData struct {
	TaskID     string       `json:"task_id"`
	RequestID  string       `json:"request_id"`
	TaskStatus string       `json:"task_status"`
	Choices    []Message    `json:"choices"`
	Usage      RequestUsage `json:"usage"`
}

type RequestUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}
