package main

import (
	"fmt"
	"glm/client"
)

func main() {
	prompt := []client.Message{
		{Role: "user", Content: "你好"},
		{Role: "assistant", Content: "我是人工智能助手"},
		{Role: "user", Content: "你叫什么名字"},
		{Role: "assistant", Content: "我叫chatGLM"},
		{Role: "user", Content: "你都可以做些什么事"},
	}
	model := "chatglm_6b"
	proxy := client.NewChatGMLClient("xxx.xxx")
	response, err := proxy.ChatComplete(model, 0.2, prompt)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Response:", response)
}
