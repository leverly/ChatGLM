package main

import (
	"fmt"
	"glm/client"
	"time"
)

var proxy *client.ChatGMLClient

func main() {
	proxy = client.NewChatGMLClient("XXX.XXX", 30*time.Second)

	prompt := []client.Message{
		{Role: "user", Content: "你好"},
		{Role: "assistant", Content: "我是人工智能助手"},
		{Role: "user", Content: "你叫什么名字"},
		{Role: "assistant", Content: "我叫chatGLM"},
		{Role: "user", Content: "你都可以做些什么事"},
	}
	model := "chatglm_6b"
	response, err := proxy.Invoke(model, 0.2, prompt)
	if err != nil {
		fmt.Println("Invoke Error:", err)
		return
	}
	fmt.Printf("Invoke Response:%s\n", (*response.Choices)[0].Content)
}

func AsyncInvoke(model string, prompt []client.Message) {
	taskId, err := proxy.AsyncInvoke(model, 0.2, prompt)
	if err != nil {
		fmt.Println("Async Invoke Error:", err)
		return
	}
	for true {
		response, err := proxy.AsyncInvokeTask(model, taskId)
		if err != nil {
			fmt.Println("Async Invoke Task Error:", err)
			return
		}
		if response.TaskStatus == client.TaskStatusSuccess {
			fmt.Printf("AsyncInvoke Response:%s\n", (*response.Choices)[0].Content)
			break
		} else if response.TaskStatus == client.TaskStatusFail {
			fmt.Println("Check Task Status Failed")
			return
		} else {
			time.Sleep(200 * time.Millisecond)
		}
	}
}
