# ChatGLM-golang
ChatGLM golang restapi sdk

# download the SDK
go get -u github.com/leverly/ChatGLM/

# client init
proxy = client.NewChatGLMClient("XXX.XXX", 30*time.Second)

# Sync Method Invoke
prompt := []client.Message{
		{Role: "user", Content: "hello world"},
}

response, err := proxy.Invoke("chatglm_6b", 0.2, prompt)

# Async Method Invoke

## submit a task
taskId, err := proxy.AsyncInvoke("chatglm_6b", 0.2, prompt)

## query the task status and result
response, err := proxy.AsyncInvokeTask("chatglm_6b", taskId)

# Stream Invoke

## callback definition
type StreamCallback struct {
}

func (s *StreamCallback) OnData(data *client.SSEInvokeResponse) {
	fmt.Print(data.Data)
}

func (s *StreamCallback) OnFinish(data *client.SSEInvokeResponse) {
	fmt.Println(data.Data)
}

## stream method invoke
err := proxy.SSEInvoke("chatglm_6b", 0.2, prompt, &StreamCallback{})
