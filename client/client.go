package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type ChatGLMClient struct {
	urlPrefix string
	apiKey    string
	client    *http.Client
}

func NewChatGLMClient(apiKey string, timeout time.Duration) *ChatGLMClient {
	return &ChatGLMClient{
		apiKey:    apiKey,
		urlPrefix: "https://open.bigmodel.cn/api/paas/v3/model-api",
		client:    &http.Client{Timeout: timeout},
	}
}

// sync invoke get result
func (c *ChatGLMClient) Invoke(model string, temperature float32, prompt []Message) (*ResponseData, error) {
	token, err := GenerateToken(c.apiKey, 300)
	if err != nil {
		return nil, err
	}
	request := InvokeRequest{
		Model:       model,
		Temperature: temperature,
		Top_p:       0.7,
		Prompt:      prompt,
	}
	requestJSON, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/%s/invoke", c.urlPrefix, model)
	req, err := http.NewRequest("POST", url, strings.NewReader(string(requestJSON)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer c.client.CloseIdleConnections()

	var response InvokeResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	} else if response.Success == false {
		return nil, fmt.Errorf("response error code:%d, msg:%s", response.Code, response.Msg)
	}
	return response.Data, nil
}

// return taskId for async invoke
func (c *ChatGLMClient) AsyncInvoke(model string, temperature float32, prompt []Message) (string, error) {
	token, err := GenerateToken(c.apiKey, 300)
	if err != nil {
		return "", err
	}

	request := InvokeRequest{
		Model:       model,
		Temperature: temperature,
		Top_p:       0.7,
		Prompt:      prompt,
	}
	requestJSON, err := json.Marshal(request)
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("%s/%s/async-invoke", c.urlPrefix, model)
	req, err := http.NewRequest("POST", url, strings.NewReader(string(requestJSON)))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var response InvokeResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", err
	} else if response.Success == false {
		return "", fmt.Errorf("response error code:%d, msg:%s", response.Code, response.Msg)
	}
	return response.Data.TaskID, nil
}

// get async invoke task result
func (c *ChatGLMClient) AsyncInvokeTask(model, taskId string) (*ResponseData, error) {
	token, err := GenerateToken(c.apiKey, 300)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/%s/async-invoke/%s", c.urlPrefix, model, taskId)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response InvokeResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	} else if response.Success == false {
		return nil, fmt.Errorf("response error code:%d, msg:%s", response.Code, response.Msg)
	}
	return response.Data, nil
}

func (c *ChatGLMClient) SSEInvoke(model string, temperature float32, prompt []Message, callback StreamEventCallback) error {
	token, err := GenerateToken(c.apiKey, 300)
	if err != nil {
		return err
	}
	request := InvokeRequest{
		Model:       model,
		Temperature: temperature,
		Top_p:       0.7,
		Prompt:      prompt,
	}
	requestJSON, err := json.Marshal(request)
	if err != nil {
		return err
	}
	url := fmt.Sprintf("%s/%s/sse-invoke", c.urlPrefix, model)
	req, err := http.NewRequest("POST", url, strings.NewReader(string(requestJSON)))
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	reader := NewEventStreamReader(resp.Body, 10240)
	for {
		event, err := reader.ReadEvent()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		err = process(event, callback)
		if err != nil {
			return err
		}
	}
	return errors.New("event stream closed")
}

func process(event *StreamEvent, callback StreamEventCallback) error {
	switch string(event.Event) {
	case AddEvent:
		callback.OnData(&SSEInvokeResponse{
			ID:   string(event.ID),
			Data: string(event.Data),
		})
	case FinishEvent:
		var task SSEResponseTaskData
		err := json.Unmarshal(event.Meta, &task)
		if err != nil {
			return err
		}
		callback.OnFinish(&SSEInvokeResponse{
			ID:   string(event.ID),
			Data: "",
			Task: &task,
		})
	case InterruptEvent:
		callback.OnInterrupt()
	case ErrorEvent:
		callback.OnError(errors.New("error"))
	default:
		return errors.New("not supported event")
	}
	return nil
}
