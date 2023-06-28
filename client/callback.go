package client

const (
	AddEvent       = "add"
	FinishEvent    = "finish"
	ErrorEvent     = "error"
	InterruptEvent = "interrupted"
)

type SSEInvokeResponse struct {
	ID    string
	Event string
	Data  string
	Task  SSEResponseTaskData
}

type SSEResponseTaskData struct {
	TaskId     string   `json:"task_id"`
	TaskStatus string   `json:"task_status"`
	RequestId  string   `json:"request_id"`
	Usage      SSEUsage `json:"usage"`
}

type SSEUsage struct {
	TotalTokens int `json:"total_tokens"`
}

type StreamEventCallback interface {
	OnData(data *SSEInvokeResponse)
	OnFinish(data *SSEInvokeResponse)
	OnError(err error)
	OnInterrupt()
}
