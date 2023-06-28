package client

import (
	"bufio"
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
	reader := bufio.NewReader(resp.Body)
	defer resp.Body.Close()

	var data, id, eventType string
	for {
		bs, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			callback.OnError(err)
			return err
		}
		if err == io.EOF {
			break
		}
		if len(bs) <= 1 {
			continue
		}
		sp := strings.Split(bs, ":")
		if len(sp) < 2 {
			err := errors.New("invalid result format")
			callback.OnError(err)
			return err
		}
		switch sp[0] {
		case "id":
			id = strings.TrimRight(sp[1], "\n")
		case "data":
			{
				temp := bs[len("data:"):]
				data = temp[:len(temp)-1]
				if eventType == AddEvent {
					callback.OnData(&SSEInvokeResponse{
						ID:    id,
						Event: eventType,
						Data:  data,
					})
				}
			}
		case "event":
			eventType = strings.TrimRight(sp[1], "\n")
			switch eventType {
			case AddEvent:
				// continue parse the data and then callback
			case FinishEvent:
				// continue parse the meta data and then callback
			case InterruptEvent:
				callback.OnInterrupt()
			case ErrorEvent:
				// TODO error callback
				callback.OnError(errors.New("error"))
			}
		case "meta":
			{
				usageData := []byte(bs[len("meta:"):])
				var task SSEResponseTaskData
				err := json.Unmarshal(usageData, &task)
				if err != nil {
					callback.OnError(err)
					return err
				}
				callback.OnFinish(&SSEInvokeResponse{
					ID:    id,
					Event: eventType,
					Data:  data,
					Task:  task,
				})
				return nil
			}
		}
	}
	return errors.New("event stream closed")
}
