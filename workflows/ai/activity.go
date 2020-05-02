package ai

import (
	"context"
	"fmt"
	"go.uber.org/cadence/activity"
	"go.uber.org/zap"
	"time"
)

/**
 * Sample activities
 */
const (
	checkNSFWActivityName  = "checkNSFWActivity"
	correctWatermarkActivityName = "correctWatermarkActivity"
)

// This is registration process where you register all your activity handlers.
func init() {
	activity.RegisterWithOptions(
		checkNSFWActivity,
		activity.RegisterOptions{Name: checkNSFWActivityName},
	)
	activity.RegisterWithOptions(
		correctWatermarkActivity,
		activity.RegisterOptions{Name: correctWatermarkActivityName},
	)
}

func checkNSFWActivity(ctx context.Context, jobID, url string) error {
	logger := activity.GetLogger(ctx)
	logger.Info("checkNSFW started", zap.String("jobID", jobID))

	fmt.Println(jobID, time.Now(), "checkNSFW Activity -> Start")

	err := checkNSFW(ctx, url)
	if err != nil {
		fmt.Println(jobID, time.Now(), CheckNSFWActivityErrorMsg)
		return err
	}

	fmt.Println(jobID, time.Now(), "checkNSFW Activity -> Finished")

	return nil
}

func correctWatermarkActivity(ctx context.Context, jobID string, url string, ) error {
	logger := activity.GetLogger(ctx)
	logger.Info("correctWatermark started.", zap.String("jobID", jobID))

	fmt.Println(jobID, time.Now(), "correctWatermark Activity -> Start")

	err := correctWatermark(ctx, url)
	if err != nil {
		fmt.Println(jobID, time.Now(), CorrectWatermarkActivityErrorMsg)
		return err
	}

	fmt.Println(jobID, time.Now(), "correctWatermark Activity -> Finished")

	return nil
}

func checkNSFW(ctx context.Context, url string) error {
	//TODO
	return nil
}

func correctWatermark(ctx context.Context, url string) error {
	//TODO
	return nil
}
