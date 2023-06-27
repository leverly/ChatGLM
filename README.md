# glm
ChatGLM golang restapi sdk

# client init
proxy = client.NewChatGMLClient("XXX.XXX", 30*time.Second)

# Sync Method Invoke
prompt := []client.Message{
		{Role: "user", Content: "你好"},
		{Role: "assistant", Content: "我是人工智能助手"},
		{Role: "user", Content: "你叫什么名字"},
		{Role: "assistant", Content: "我叫chatGLM"},
		{Role: "user", Content: "你都可以做些什么事"},
}

response, err := proxy.Invoke("chatglm_6b", 0.2, prompt)

# Async Method Invoke

## submit a task
taskId, err := proxy.AsyncInvoke("chatglm_6b", 0.2, prompt)

## query the task status and result
response, err := proxy.AsyncInvokeTask("chatglm_6b", taskId)
