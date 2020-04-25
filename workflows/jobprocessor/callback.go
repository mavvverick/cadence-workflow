package jobprocessor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

//CallbackInfo ...
type CallbackInfo struct {
	URL       string
	Status    string
	Type      string
	TaskToken string
	Event     string
	Payload   string
}

//NewCallbackInfo ...
func NewCallbackInfo(format *Format) *CallbackInfo {
	return &CallbackInfo{
		URL:     format.CallbackURL,
		Payload: format.Payload,
	}
}

type webhookMessage struct {
	Status       string `json:"status"`
	CallbackType string `json:"callback_type"`
	TaskToken    string `json:"task_token"`
	Event        string `json:"event"`
	Payload      string `json:"payload"`
	ErrorCode    int    `json:"error_code,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
}

//PushMessage ...
func (e *CallbackInfo) PushMessage(status, callbackType, token, event string) {
	requestBody := &webhookMessage{
		Status:       fmt.Sprintf(`{"status":"%v"}`, status),
		CallbackType: callbackType,
		TaskToken:    token,
		Event:        event,
		Payload:      e.Payload,
	}

	if event == "error" {
		requestBody.ErrorCode = 500
		requestBody.ErrorMessage = status
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Println("err")
	}
	req, err := http.NewRequest("POST", e.URL, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()
}
