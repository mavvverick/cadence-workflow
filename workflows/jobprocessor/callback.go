package jobprocessor

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	ka "github.com/YOVO-LABS/workflow/common/messaging"
	"go.uber.org/cadence/activity"
)

//CallbackInfo ...
type CallbackInfo struct {
	URL       string
	Status    string
	Type      string
	TaskToken string
	Event     string
	Payload   string
	PostID    string
	Bucket    string
}

//NewCallbackInfo ...
func NewCallbackInfo(format *Format, bucket string) *CallbackInfo {
	return &CallbackInfo{
		URL:     format.CallbackURL,
		PostID:  strings.Split(format.Payload, "|")[0],
		Payload: format.Payload,
		Bucket:  bucket,
	}
}

type webhookMessage struct {
	Status       string `json:"status"`
	TaskToken    string `json:"taskToken"`
	Event        string `json:"event"`
	PostID       string `json:"postID"`
	ErrorCode    string `json:"errorCode,omitempty"`
	ErrorMessage string `json:"errorMessage,omitempty"`
}

//PushMessage ...
func (e *CallbackInfo) PushMessage(ctx context.Context, status, callbackType, token, event string) {

	if activity.GetInfo(ctx).Attempt < 1 && (event == CallbackErrorEvent || event == CallbackRejectEvent) {
		return
	}
	requestBody := &webhookMessage{
		Status:    status,
		TaskToken: token,
		Event:     event,
		PostID:    e.Payload,
	}

	if event == "error" {
		requestBody.ErrorCode = "500"
		requestBody.ErrorMessage = status
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Println("err")
	}

	kafkaClientName := "kafkaCallbackClient"
	if e.Bucket != "yovo-app" {
		kafkaClientName = "kafkaDevCallbackClient"
	}

	kafkaClient := ctx.Value(kafkaClientName).(ka.KafkaAdapter)
	if body != nil {
		err = kafkaClient.Producer.Publish(ctx, e.Payload, string(body))
		if err != nil {
			fmt.Println("Cannot push message to kafka")
		}
	}

	// req, err := http.NewRequest("POST", e.URL, bytes.NewBuffer(body))
	// req.Header.Set("Content-Type", "application/json")

	// client := &http.Client{}
	// resp, err := client.Do(req)
	// if err != nil {
	// 	panic(err)
	// }

	// defer resp.Body.Close()
}
