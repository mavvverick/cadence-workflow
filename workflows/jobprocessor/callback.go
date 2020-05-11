package jobprocessor

import (
	"context"
	"encoding/json"
	"fmt"
	"go.uber.org/cadence/activity"
	"os"
	"strings"

	ka "github.com/YOVO-LABS/workflow/common/messaging"
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
		Payload: strings.Split(format.Payload, "|")[0],
	}
}

type webhookMessage struct {
	Status       string `json:"status"`
	TaskToken    string `json:"taskToken"`
	Event        string `json:"event"`
	PostID       string `json:"postID"`
	ErrorCode    int    `json:"errorCode,omitempty"`
	ErrorMessage string `json:"errorMessage,omitempty"`
}

//PushMessage ...
func (e *CallbackInfo) PushMessage(ctx context.Context, status, callbackType, token, event string) {
	if activity.GetInfo(ctx).Attempt < 1 && event == CallbackErrorEvent {
		return
	}
	requestBody := &webhookMessage{
		Status:    fmt.Sprintf(`{"status":"%v"}`, status),
		TaskToken: token,
		Event:     event,
		PostID:    e.Payload,
	}

	if event == "error" {
		requestBody.ErrorCode = 500
		requestBody.ErrorMessage = status
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Println("err")
	}
	kafkaClient := ctx.Value("kafkaClient").(ka.KafkaAdapter)

	if body != nil {
		err = kafkaClient.Producer.Publish(os.Getenv("CB_TOPIC"), "video", string(body))
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
