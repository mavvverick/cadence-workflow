package jobprocessor

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"jobprocessor/api/model"

	"cloud.google.com/go/storage"
	"go.uber.org/cadence/activity"
	"go.uber.org/zap"
)

/**
 * Sample activities used by file processing sample workflow.
 */
const (
	createJobActivityName         = "createJobActivity"
	waitForDecisionActivityName   = "waitForDecisionActivity"
	downloadFileActivityName      = "downloadFileActivity"
	compressFileActivityName      = "compressFileActivity"
	uploadFileActivityName        = "uploadFileActivity"
	migrateToColdLineActivityName = "migrateToColdLineActivity"
	srcDirectory                  = "raw/"
	processedDirectory            = "processed/"
	blackHole                     = "blackHole/"
	localDirectory                = "/Users/arpit/go/src/jobprocessor/videos/raw/"
	localProcessedDirectory       = "/Users/arpit/go/src/jobprocessor/videos/processed/"
)

// This is registration process where you register all your activity handlers.
func init() {
	activity.RegisterWithOptions(
		createJobActivity,
		activity.RegisterOptions{Name: createJobActivityName},
	)
	activity.RegisterWithOptions(
		waitForDecisionActivity,
		activity.RegisterOptions{Name: waitForDecisionActivityName},
	)
	activity.RegisterWithOptions(
		downloadFileActivity,
		activity.RegisterOptions{Name: downloadFileActivityName},
	)
	activity.RegisterWithOptions(
		compressFileActivity,
		activity.RegisterOptions{Name: compressFileActivityName},
	)
	activity.RegisterWithOptions(
		uploadFileActivity,
		activity.RegisterOptions{Name: uploadFileActivityName},
	)
	activity.RegisterWithOptions(
		migrateToColdLineActivity,
		activity.RegisterOptions{Name: migrateToColdLineActivityName},
	)
}

func createJobActivity(ctx context.Context, jobID string) error {
	if len(jobID) == 0 {
		return errors.New("job id is empty")
	}

	resp, err := http.Get(jobServerURL + "/cadence/job/create?is_api_call=true&id=" + jobID)

	if err != nil {
		return err
	}
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return err
	}

	if string(body) == "SUCCEED" {
		activity.GetLogger(ctx).Info("Job created.", zap.String("JobID", jobID))
		return nil
	}

	return errors.New(string(body))
}

func waitForDecisionActivity(ctx context.Context, jobID string) (string, error) {
	if len(jobID) == 0 {
		return "", errors.New("job id is empty")
	}

	logger := activity.GetLogger(ctx)

	activityInfo := activity.GetInfo(ctx)
	formData := url.Values{}
	formData.Add("task_token", string(activityInfo.TaskToken))

	registerCallbackURL := jobServerURL + "/cadence/job/register?id=" + jobID
	resp, err := http.PostForm(registerCallbackURL, formData)
	if err != nil {
		logger.Info("waitForDecisionActivity failed to register callback.", zap.Error(err))
		return "", err
	}
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return "", err
	}

	status := string(body)
	if status == "SUCCEED" {
		// register callback succeed
		logger.Info("Successfully registered callback.", zap.String("jobID", jobID))

		// ErrActivityResultPending is returned from activity's execution to indicate the activity is not
		//completed when it returns. Activity will be completed asynchronously when Client.CompleteActivity() is called.
		return "", activity.ErrResultPending
	}

	logger.Warn("Register callback failed.", zap.String("job Status", status))
	return "", fmt.Errorf("register callback failed status:%s", status)
}

func downloadFileActivity(ctx context.Context, jobID, url string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Downloading file...", zap.String("File URL", url))

	fpath, err := downloadFile(ctx, url)

	if err != nil {
		return "", err
	}
	return fpath, nil
}

func compressFileActivity(ctx context.Context, jobID string, filepath string, format model.Format) (string, error) {
	var compressFlag string
	logger := activity.GetLogger(ctx).With(zap.String("HostID", HostID))
	logger.Info("processFileActivity started.", zap.String("FileName", filepath))

	// process the file
	fname, err := compressFile(ctx, filepath, format)

	if err != nil {
		logger.Error("processFileActivity failed to rename file.", zap.Error(err))
		compressFlag := "FAILED"
		return compressFlag, err
	}

	logger.Info("processFileActivity succeed.", zap.String("SavedFilePath", fname))
	compressFlag = "SUCCESS"
	return compressFlag, nil
}

func uploadFileActivity(ctx context.Context, jobID string, format model.Format) error {
	logger := activity.GetLogger(ctx).With(zap.String("HostID", HostID))
	logger.Info("uploadFileActivity begin", zap.String("FileName", localProcessedDirectory))

	// upload the file
	err := uploadFile(ctx, format)
	if err != nil {
		return err
	}
	logger.Info("uploadFileActivity succeeded", zap.String("FileName", localProcessedDirectory))
	return nil
}

func migrateToColdLineActivity(ctx context.Context, jobID string, encode model.Encode) error {
	logger := activity.GetLogger(ctx).With(zap.String("HostID", HostID))
	logger.Info("migrateToColdline begin...", zap.String("file to move", encode.Source))

	err := migrateToColdline(ctx, jobID, encode)
	if err != nil {
		logger.Error("migrateToColdline uploading failed.", zap.Error(err))
		return err
	}
	return nil
}

func downloadFile(ctx context.Context, url string) (string, error) {

	bucket := strings.Split(strings.Split(url, ".")[0], "//")[1]
	object := strings.Split(url, ".com/")[1]
	localFileName := localDirectory + strings.Split(object, "/")[2]

	err := downloadObjectToLocal(bucket, object, localFileName)
	if err != nil {
		return "", err
	}

	return localFileName, nil
}

func compressFile(ctx context.Context, filepath string, format model.Format) (string, error) {
	// Two pass encoding
	encodeCmdPass0 := createEncodeCommand(filepath, 0, format.Encode)
	fmt.Println(encodeCmdPass0)
	argsPass0 := strings.Fields(encodeCmdPass0)
	cmdPass0 := exec.Command(argsPass0[0], argsPass0[1:]...)
	errPass0 := executeEncodeCommand(ctx, cmdPass0)
	if cmdPass0 != nil {
		if errPass0 != nil {
			return "", errPass0
		}
	}

	encodeCmdPass1 := createEncodeCommand(filepath, 1, format.Encode)
	fmt.Println(encodeCmdPass1)
	argsPass1 := strings.Fields(encodeCmdPass1)
	cmdPass1 := exec.Command(argsPass1[0], argsPass1[1:]...)
	errPass1 := executeEncodeCommand(ctx, cmdPass1)
	if errPass1 != nil {
		return "", errPass1
	}
	return "", nil
}

func createEncodeCommand(filepath string, pass int, encodes []model.Encode) string {
	encodeCmd := "ffmpeg" + " -i " + filepath

	for _, encode := range encodes {
		pixelFormat := "yuv420p"
		rate := strconv.Itoa(encode.BitRate) + "k"

		videoCodec, framerate, tagVideo := encode.VideoCodec, encode.FrameRate, "hvc1"
		bitRate, bufferSize, maxRate := rate, rate, rate
		preset, videoFormat := "medium", encode.VideoFormat
		outputPath := localProcessedDirectory + encode.VideoCodec + "_" + encode.Size

		encodeCmd = encodeCmd +
			" -pix_fmt " + pixelFormat +
			" -vsync " + "1" +
			" -vcodec " + videoCodec +
			" -r " + strconv.Itoa(framerate) +
			" -threads " + "0" +
			" -b:v: " + bitRate +
			" -bufsize " + bufferSize +
			" -maxrate " + maxRate +
			" -preset " + preset +
			" -f " + videoFormat +
			" -pass " + strconv.Itoa(pass)

		if videoCodec == "libx265" {
			encodeCmd = encodeCmd + " -tag:v " + tagVideo + " -y " + outputPath
		}
	}

	return encodeCmd
}

func executeEncodeCommand(ctx context.Context, cmd *exec.Cmd) error {
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		logger := activity.GetLogger(ctx).With(zap.String("HostID", HostID))
		logger.Info("Error running the command" + ": " + stderr.String())
		return err
	}
	return nil
}

func uploadFile(ctx context.Context, format model.Format) error {
	gsContext := context.Background()
	storageClient, err := storage.NewClient(gsContext)
	if err != nil {
		return err
	}
	for _, encode := range format.Encode {
		bucket := strings.Split(strings.Split(encode.Destination, ".")[0], "//")[1]
		object := strings.Split(encode.Destination, ".com/")[1]
		filepath := localProcessedDirectory + encode.VideoCodec + "_" + encode.Size
		file, err := os.Open(filepath)
		writeContext := storageClient.Bucket(bucket).Object(object).NewWriter(gsContext)
		if _, err = io.Copy(writeContext, file); err != nil {
			return err
		}
		if err := writeContext.Close(); err != nil {
			return err
		}
		defer file.Close()

	}
	return nil
}

func migrateToColdline(ctx context.Context, jobID string, encode model.Encode) error {
	gsContext := context.Background()

	gsContext, cancel := context.WithTimeout(gsContext, time.Second*10)
	defer cancel()

	storageClient, err := storage.NewClient(gsContext)

	if err != nil {
		return nil
	}

	bucket := strings.Split(strings.Split(encode.Source, ".")[0], "//")[1]
	gcsObject := strings.Split(encode.Source, ".com/")[1]

	src := storageClient.Bucket(bucket).Object(gcsObject)
	dst := storageClient.Bucket(bucket).Object(blackHole + gcsObject)

	if _, err := dst.CopierFrom(src).Run(gsContext); err != nil {
		return err
	}
	if err := src.Delete(gsContext); err != nil {
		return err
	}

	resp, err := http.Get(jobServerURL + "/cadence/job/action?is_api_call=true&type=processed&id=" + jobID)
	if err != nil {
		return err
	}
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return err
	}

	if string(body) == "SUCCEED" {
		activity.GetLogger(ctx).Info("migrateToColdline succeed", zap.String("jobID", jobID))
		return nil
	}

	return errors.New(string(body))
}

func downloadObjectToLocal(bucket, object, localDirectory string) error {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}

	readContext, err := client.Bucket(bucket).Object(object).NewReader(ctx)
	if err != nil {
		return err
	}

	defer readContext.Close()

	data, err := ioutil.ReadAll(readContext)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(localDirectory, data, 0644)
	if err != nil {
		return err
	}

	return nil
}
