# glm
ChatGLM golang restapi sdk

# client init
proxy = client.NewChatGLMClient("XXX.XXX", 30*time.Second)

# Sync Method Invoke
prompt := []client.Message{
		{Role: "user", Content: "你好"},
}

response, err := proxy.Invoke("chatglm_6b", 0.2, prompt)

# Async Method Invoke

## submit a task
taskId, err := proxy.AsyncInvoke("chatglm_6b", 0.2, prompt)

## query the task status and result
response, err := proxy.AsyncInvokeTask("chatglm_6b", taskId)

# Stream Invoke

## callback definition

func (s *StreamCallback) OnData(data *client.SSEInvokeResponse) {
	fmt.Print(data.Data)
}

func (s *StreamCallback) OnFinish(data *client.SSEInvokeResponse) {
	fmt.Println(data.Data)
}

## Stream Method Invoke
err := proxy.SSEInvoke(model, 0.2, prompt, &callback)
