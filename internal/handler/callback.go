package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/YOVO-LABS/workflow/api/model"
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
func NewCallbackInfo(url string) *CallbackInfo {
	return &CallbackInfo{
		URL: url,
	}
}

//PushMessage ...
func (e *CallbackInfo) PushMessage(status, callbackType, token, event string, payload []model.Encode) {
	pl, err := json.Marshal(&payload)
	if err != nil {
		panic(err)
	}

	requestBody := fmt.Sprintf(
		`{"status":{"status": "%v", "callback_type": "%v", "task_token": "%v", "event": "%v", "payload": %v}}`,
		status,
		callbackType,
		token,
		event,
		string(pl))

	jsonStr := []byte(requestBody)
	req, err := http.NewRequest("POST", e.URL, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()
}
