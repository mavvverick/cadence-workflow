package ai

import (
	"context"
	"fmt"
	"time"

	"github.com/YOVO-LABS/workflow/proto/dense"

	"go.uber.org/cadence/activity"
	"go.uber.org/zap"
)

/**
 * Sample activities
 */
const (
	checkNSFWActivityName        = "checkNSFWActivity"
	correctWatermarkActivityName = "correctWatermarkActivity"
)

// This is registration process where you register all your activity handlers.
func init() {
	activity.RegisterWithOptions(
		checkNSFWAndLogoActivity,
		activity.RegisterOptions{Name: checkNSFWActivityName},
	)
}

func checkNSFWAndLogoActivity(ctx context.Context, jobID, url string) error {
	logger := activity.GetLogger(ctx)
	logger.Info("checkNSFW started", zap.String("jobID", jobID))

	fmt.Println(jobID, time.Now(), "checkNSFW Activity -> Start")

	err := checkNSFWAndLogo(ctx, url)
	if err != nil {
		fmt.Println(jobID, time.Now(), CheckNSFWActivityErrorMsg)
		return err
	}

	fmt.Println(jobID, time.Now(), "checkNSFW Activity -> Finished")
	return nil
}

func checkNSFWAndLogo(ctx context.Context, url string) error {
	mlClient := ctx.Value("mlClient").(dense.PredictClient)
	imageData := &dense.ImageData{
		PostId: url,
	}
	predictResponse, err := mlClient.PredictPipeline(ctx, imageData)
	if err != nil {
		return err
	}
	fmt.Println(predictResponse)
	return nil
}
