package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/YOVO-LABS/workflow/api/model"
	jp "github.com/YOVO-LABS/workflow/workflows/jobprocessor"
	lb "github.com/YOVO-LABS/workflow/workflows/leaderboard"

	adapter "github.com/YOVO-LABS/workflow/internal/adapter"

	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/cadence/client"
	"go.uber.org/zap"
)

//Service ..
type Service struct {
	CadenceAdapter *adapter.CadenceAdapter
	// KafkaAdapter   *adapter.KafkaAdapter
	Logger *zap.Logger
}

//NewService ...
func NewService(cadenceAdapter *adapter.CadenceAdapter, Logger *zap.Logger) Service {
	return Service{cadenceAdapter, Logger}
}

//TriggerJobProcessWorkflow ...
func (b *Service) TriggerJobProcessWorkflow(responseWriter http.ResponseWriter, request *http.Request) {

	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		http.Error(responseWriter, err.Error(), http.StatusBadRequest)
		return
	}

	var q model.Query
	err = json.Unmarshal(body, &q)
	if err != nil {
		http.Error(responseWriter, err.Error(), http.StatusBadRequest)
		return
	}

	var encodes []model.Encode
	for _, format := range q.Format {
		mp4EncodeParams := model.NewMP4Encode()
		mp4EncodeParams.
			SetDestination(format.Destination.URL).
			SetSize(format.Size).
			SetVideoCodec(format.VideoCodec).
			SetFrameRate(format.Framerate).
			SetBitRate(format.Bitrate).
			SetBufferSize(format.Bitrate).
			SetMaxRate(format.Bitrate).
			SetVideoFormat(format.FileExtension).
			GetEncode()
		encodes = append(encodes, mp4EncodeParams.GetEncode())
	}

	videoFormat := model.NewVideoFormat()
	videoFormat.
		SetFormatSource(q.Source).
		SetFormatEncode(encodes).
		GetFormat()

	fmt.Println(videoFormat.GetFormat().Source)

	workflowOptions := client.StartWorkflowOptions{
		ID:                              "jobProcessing_" + uuid.New().String(),
		TaskList:                        jp.TaskList,
		ExecutionStartToCloseTimeout:    time.Minute,
		DecisionTaskStartToCloseTimeout: time.Minute,
	}

	execution, err := b.CadenceAdapter.CadenceClient.StartWorkflow(
		context.Background(),
		workflowOptions,
		jp.Workflow,
		uuid.New().String(),
		videoFormat.GetFormat(),
	)

	if err != nil {
		http.Error(responseWriter, "Error starting workflow again! "+err.Error(), http.StatusBadRequest)
		return
	}
	js, _ := json.Marshal(execution)
	responseWriter.Header().Set("Content-Type", "application/json")
	_, _ = responseWriter.Write(js)
}

//TriggerCronLeaderBoardProcessWorkflow ...
func (b *Service) TriggerCronLeaderBoardProcessWorkflow(responseWriter http.ResponseWriter, request *http.Request) {

	//parse request body to get cronTime
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		http.Error(responseWriter, err.Error(), http.StatusBadRequest)
		return
	}

	// fmt.Println("payload")

	var payload map[string]interface{}
	err = json.Unmarshal(body, &payload)
	if err != nil {
		http.Error(responseWriter, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Println(payload)

	cronSchedule := strings.Split(payload["cronTime"].(string), ":")

	cronTime := fmt.Sprintf("%s %s %s %s %s", cronSchedule[1], cronSchedule[0], "*", "*", "*")

	fmt.Println(cronTime)
	workflowOptions := client.StartWorkflowOptions{
		ID:                              "leaderBoardProcess_" + uuid.New().String(),
		TaskList:                        lb.TaskList,
		ExecutionStartToCloseTimeout:    time.Hour * 2,
		DecisionTaskStartToCloseTimeout: time.Hour * 2,
		CronSchedule:                    cronTime,
	}

	execution, err := b.CadenceAdapter.CadenceClient.StartWorkflow(
		context.Background(),
		workflowOptions,
		lb.Workflow,
		uuid.New().String(),
		// payload["timer"],
	)

	if err != nil {
		http.Error(responseWriter, "Error starting workflow again! "+err.Error(), http.StatusBadRequest)
		return
	}
	js, _ := json.Marshal(execution)
	responseWriter.Header().Set("Content-Type", "application/json")
	_, _ = responseWriter.Write(js)
}

// UpdateCronLeaderBoardProcessWorkflow ...
func (b *Service) UpdateCronLeaderBoardProcessWorkflow(responseWriter http.ResponseWriter, request *http.Request) {
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		http.Error(responseWriter, err.Error(), http.StatusBadRequest)
		return
	}
	req, _ := http.NewRequest("POST", "http://localhost:3030/cadence/cron/terminate", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	body, _ = ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))
}

//TerminateWorkflow ...
func (b *Service) TerminateWorkflow(responseWriter http.ResponseWriter, request *http.Request) {
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		http.Error(responseWriter, err.Error(), http.StatusBadRequest)
		return
	}

	var payload map[string]string
	err = json.Unmarshal(body, &payload)
	if err != nil {
		http.Error(responseWriter, err.Error(), http.StatusBadRequest)
		return
	}

	err = b.CadenceAdapter.CadenceClient.TerminateWorkflow(
		context.Background(),
		payload["workflowId"],
		"",
		"",
		nil,
	)

	if err != nil {
		http.Error(responseWriter, "Error terminating workflow again! "+err.Error(), http.StatusBadRequest)
		return
	}
	js, _ := json.Marshal("")
	responseWriter.Header().Set("Content-Type", "application/json")
	_, _ = responseWriter.Write(js)
}

// ListCronProcessWorkflow .....

// func (b  *Service) ListCronProcessWorkflow(responseWriter http.ResponseWriter, request *http.Request) {
// 	body, err := ioutil.ReadAll(request.Body)

// 	if err != nil {
// 		http.Error(responseWriter, err.Error(), http.StatusBadRequest)
// 		return
// 	}

// 	// var q model.
// 	var payload map[string]interface{}
// 	err = json.Unmarshal(body, &q)
// 	if err != nil {
// 		http.Error(responseWriter, err.Error(), http.StatusBadRequest)
// 		return
// 	}

// 	listOpenWorkflowExecutionsRequest :=  {
// 		Domain:

// 	}

// 	execution, err := b.CadenceAdapter.CadenceClient.ListWorkflow(
// 		context.Background(),
// 		shared.
// 	)

// 	if err != nil {
// 		http.Error(responseWriter, "Error starting workflow again! "+err.Error(), http.StatusBadRequest)
// 		return
// 	}
// 	js, _ := json.Marshal(execution)
// 	responseWriter.Header().Set("Content-Type", "application/json")
// 	_, _ = responseWriter.Write(js)
// }

//CancelTimer ...
// func (b *Service) CancelTimer(responseWriter http.ResponseWriter, request *http.Request) {
// 	body, err := ioutil.ReadAll(request.Body)

// 	if err != nil {
// 		http.Error(responseWriter, err.Error(), http.StatusBadRequest)
// 		return
// 	}

// 	var payload map[string]string
// 	err = json.Unmarshal(body, &payload)
// 	if err != nil {
// 		http.Error(responseWriter, err.Error(), http.StatusBadRequest)
// 		return
// 	}

// 	workflowRun := b.CadenceAdapter.CadenceClient.(
// 		context.Background(),
// 		payload["workflowID"],
// 		"")

// 	// ctx, _ := context.WithTimeout(context.Background(), 90*time.Second)
// 	ctx, _ := context.WithDeadline(context.Background(), time.Now())
// 	// ctx, _ := context.WithDeadline
// 	_, err = lb.CancelTimer(ctx, payload["workflowID"])

// 	if err != nil {
// 		http.Error(responseWriter, "Error cancelling timer again! "+err.Error(), http.StatusBadRequest)
// 		return
// 	}
// 	js, _ := json.Marshal("Timer cancelled")
// 	responseWriter.Header().Set("Content-Type", "application/json")
// 	_, _ = responseWriter.Write(js)

// }
