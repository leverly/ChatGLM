package main

import (
	"fmt"
	"glm/client"
)

type StreamCallback struct {
}

func (s *StreamCallback) OnData(data *client.SSEInvokeResponse) {
	fmt.Print(data.Data)
}

func (s *StreamCallback) OnFinish(data *client.SSEInvokeResponse) {
	fmt.Println(data.Data)
}

func (s *StreamCallback) OnError(err error) {
	fmt.Println("error:", err)
}

func (s *StreamCallback) OnInterrupt() {
	fmt.Println("interrupted")
}
