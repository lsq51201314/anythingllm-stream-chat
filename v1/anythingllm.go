package anythingllm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type reqBody struct {
	Message string `json:"message"`
	Mode    string `json:"mode"`
	Stream  bool   `json:"stream"`
	Reset   bool   `json:"reset"`
}

type repInfo struct {
	UUID         string `json:"uuid"`
	Type         string `json:"type"`
	TextResponse string `json:"textResponse"`
	Close        bool   `json:"close"`
	Error        bool   `json:"error"`
}

// 流式回调
type ChatCallback func(uuid string, message string)

type AnythingLLM struct {
	url           string
	authorization string
	callback      ChatCallback
}

// 新建实例
func New(ip string, port int, slug string, authorization string) *AnythingLLM {
	obj := AnythingLLM{
		url:           fmt.Sprintf("http://%s:%d/api/v1/workspace/%s/stream-chat", ip, port, slug),
		authorization: authorization,
	}
	return &obj
}

// 绑定回调
func (a *AnythingLLM) SetChatCallback(cfunc ChatCallback) {
	a.callback = cfunc
}

// 发送聊天
func (a *AnythingLLM) StreamChat(message string, reset ...bool) (string, error) {
	rs := false
	if len(reset) > 0 {
		rs = reset[0]
	}
	obj := reqBody{
		Message: message,
		Mode:    "chat",
		Stream:  true,
		Reset:   rs,
	}
	data, err := json.Marshal(&obj)
	if err != nil {
		return "", err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", a.url, bytes.NewBuffer(data))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+a.authorization)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("http status:%d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType != "text/event-stream" {
		return "", fmt.Errorf("content type:%s", contentType)
	}

	scanner := bufio.NewScanner(resp.Body)
	str := ""
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || len(line) <= 5 {
			continue
		}
		line = line[5:]

		var info repInfo
		if err := json.Unmarshal([]byte(line), &info); err != nil {
			return "", err
		}

		if info.Type == "textResponseChunk" {
			if a.callback != nil {
				a.callback(info.UUID, info.TextResponse)
			}
			str += info.TextResponse
		}
	}
	return str, nil
}
