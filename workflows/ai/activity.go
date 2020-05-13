package ai

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/cadence"

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

func checkNSFWAndLogoActivity(ctx context.Context, jobID, url string, cb *jp.CallbackInfo) (*dense.Response, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("checkNSFW started", zap.String("jobID", jobID))

	fmt.Println(jobID, time.Now(), "checkNSFW Activity -> Start")
	res, err := checkNSFWAndLogo(ctx, url)
	if err != nil {
		fmt.Println(jobID, time.Now(), CheckNSFWActivityErrorMsg)
		cb.PushMessage(ctx, err.Error(), jp.Task, jobID, jp.CallbackErrorEvent)
		return nil, cadence.NewCustomError(err.Error())
	}
	if res.IsNext == false {
		fmt.Println(jobID, time.Now(), CheckNSFWActivityErrorMsg)
		if res.Error != "" {
			cb.PushMessage(ctx, res.Error, jp.Task, jobID, jp.CallbackErrorEvent)
			return nil, cadence.NewCustomError(res.Error)
		}
		cb.PushMessage(ctx, NSFWErrorMessage, jp.Task, jobID, jp.CallbackRejectEvent)
		return nil, errors.New(NSFWErrorMessage)
	}

	fmt.Println(jobID, time.Now(), "checkNSFW Activity -> Finished")
	return res, nil
}

func checkNSFWAndLogo(ctx context.Context, url string) (*dense.Response, error) {
	mlClient := ctx.Value("mlClient").(dense.PredictClient)
	imageData := &dense.ImageData{
		PostId: url,
	}
	predictResponse, err := mlClient.PredictPipeline(ctx, imageData)
	if err != nil {
		return nil, err
	}
	return predictResponse, nil
}
