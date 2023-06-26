package client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type ChatGMLClient struct {
	urlPrefix string
	apiKey    string
	client    *http.Client
}

func NewChatGMLClient(apiKey string) *ChatGMLClient {
	return &ChatGMLClient{
		apiKey:    apiKey,
		urlPrefix: "https://open.bigmodel.cn/api/paas/v3/model-api/",
		client:    &http.Client{},
	}
}

// sync method
func (c *ChatGMLClient) ChatComplete(model string, temperature float32, prompt []Message) (*ResponseData, error) {
	token, err := GenerateToken(c.apiKey, 300)
	if err != nil {
		return nil, err
	}

	request := InvokeRequest{
		Model:       model,
		Temperature: temperature,
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

	c.client.Timeout = 30 * time.Second
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
