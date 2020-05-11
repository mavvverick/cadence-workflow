package ai

import (
	"context"
	"errors"
	"fmt"
	"go.uber.org/cadence"
	"time"

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
	res, err := checkNSFWAndLogo(ctx, jobID, url, cb)
	if err != nil {
		fmt.Println(jobID, time.Now(), CheckNSFWActivityErrorMsg)
		cb.PushMessage(ctx, NSFWErrorMessage, jp.Task, jobID, jp.CallbackErrorEvent)
		return nil, err
	}

	fmt.Println(jobID, time.Now(), "checkNSFW Activity -> Finished")
	return res, nil
}

func checkNSFWAndLogo(ctx context.Context, jobID, url string, cb *jp.CallbackInfo) (*dense.Response, error) {
	mlClient := ctx.Value("mlClient").(dense.PredictClient)
	imageData := &dense.ImageData{
		PostId: url,
	}
	predictResponse, err := mlClient.PredictPipeline(ctx, imageData)
	if err != nil {
		return nil, err
	}
	if predictResponse.IsNext == false {
		if predictResponse.Error != "" {
			return nil, cadence.NewCustomError(predictResponse.Error, predictResponse.Error)
		}
		return nil, errors.New(NSFWErrorMessage)
	}
	return predictResponse, nil
}
