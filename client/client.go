package client

import (
	"encoding/json"
	"fmt"
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
