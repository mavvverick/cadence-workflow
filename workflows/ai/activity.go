package ai

import (
	"context"
	"fmt"
	"net"
	"time"

	"go.uber.org/cadence"

	"github.com/YOVO-LABS/workflow/common/monitoring"
	"github.com/YOVO-LABS/workflow/proto/dense"
	jp "github.com/YOVO-LABS/workflow/workflows/jobprocessor"

	"go.uber.org/cadence/activity"
	"go.uber.org/zap"
)

/**
 * Sample activities
 */
const (
	checkNSFWActivityName        = "checkNSFWActivity"
	correctWatermarkActivityName = "correctWatermarkActivity"
	NSFWErrorMessage             = "NSFW Content"
)

// This is registration process where you register all your activity handlers.
func init() {
	activity.RegisterWithOptions(
		checkNSFWAndLogoActivity,
		activity.RegisterOptions{Name: checkNSFWActivityName},
	)
}

func checkNSFWAndLogoActivity(ctx context.Context, jobID, postID, bucket string, cb *jp.CallbackInfo) (*dense.Response, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("checkNSFW started", zap.String("jobID", jobID))

	fmt.Println(jobID, time.Now(), "checkNSFW Activity -> Start")
	res, err := checkNSFWAndLogo(ctx, postID, bucket)
	if err != nil {
		fmt.Println(jobID, time.Now(), CheckNSFWActivityErrorMsg)
		cb.PushMessage(ctx, err.Error(), jp.Task, jobID, jp.CallbackErrorEvent)
		return nil, cadence.NewCustomError(err.Error())
	}

	if res.IsNext == false {
		if res.Error != "" {
			cb.PushMessage(ctx, res.Error, jp.Task, jobID, jp.CallbackErrorEvent)
			return nil, cadence.NewCustomError(res.Error)
		}
		ev := monitoring.AIEvent{
			PostID:  postID,
			Meta:    res.Message,
			IsTrue:  res.IsNext,
			Version: "1",
		}
		data := ev.Message()
		udpConn, ok := ctx.Value("udpConn").(*net.UDPConn)
		if !ok {
			return nil, cadence.NewCustomError("No udp connection")
		}
		monitoring.FireEvent(udpConn, data)
		// cb.PushMessage(ctx, res.Message, jp.Task, jobID, jp.CallbackRejectEvent)
		// return nil, errors.New(res.Message)
	}

	fmt.Println(jobID, time.Now(), "checkNSFW Activity -> Finished")
	return res, nil
}

func checkNSFWAndLogo(ctx context.Context, postID, bucket string) (*dense.Response, error) {
	mlClient := ctx.Value("mlClient").(dense.PredictClient)
	imageData := &dense.ImageData{
		PostId: postID,
		Bucket: bucket,
	}
	predictResponse, err := mlClient.PredictPipeline(ctx, imageData)
	if err != nil {
		return nil, err
	}
	return predictResponse, nil
}
