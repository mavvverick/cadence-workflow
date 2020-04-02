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

	"github.com/YOVO-LABS/workflow/api/model"
	"google.golang.org/api/option"

	path "path/filepath"

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
	localDirectory                = "/tmp/"
	localProcessedDirectory       = "/tmp/"
)

// This is registration process where you register all your activity handlers.
func init() {
	// activity.RegisterWithOptions(
	// 	createJobActivity,
	// 	activity.RegisterOptions{Name: createJobActivityName},
	// )

	// activity.RegisterWithOptions(
	// 	waitForDecisionActivity,
	// 	activity.RegisterOptions{Name: waitForDecisionActivityName},
	// )

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

	// activity.RegisterWithOptions(
	// 	migrateToColdLineActivity,
	// 	activity.RegisterOptions{Name: migrateToColdLineActivityName},
	// )
}

func createJobActivity(ctx context.Context, jobID string) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Inside createJobActivity")
	if len(jobID) == 0 {
		return errors.New("job id is empty")
	}

	resp, err := http.Get(jobServerURL + "/workflow/job/start?is_api_call=true&id=" + jobID)
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

	registerCallbackURL := jobServerURL + "/workflow/job/register?id=" + jobID
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

func compressFileActivity(ctx context.Context, jobID string, filepath string, format model.Format) error {
	// var compressFlag string
	logger := activity.GetLogger(ctx).With(zap.String("HostID", HostID))
	logger.Info("compressFileActivity started.", zap.String("FileName", filepath))

	// process the file
	err := compressFile(ctx, filepath, format)

	if err != nil {
		logger.Error("compressFileActivity failed to compress file.", zap.Error(err))
		// compressFlag := "FAILED"
		return err
	}

	logger.Info("compressFileActivity succeed.")
	// compressFlag = "SUCCESS"
	return nil
}

func uploadFileActivity(ctx context.Context, jobID, fpath string, format model.Format) error {
	logger := activity.GetLogger(ctx).With(zap.String("HostID", HostID))
	logger.Info("uploadFileActivity begin", zap.String("FileName", localProcessedDirectory))

	// upload the file
	err := uploadFile(ctx, fpath, format)
	if err != nil {
		return err
	}
	logger.Info("uploadFileActivity succeeded", zap.String("FileName", localProcessedDirectory))
	return nil
}

func downloadFile(ctx context.Context, url string) (string, error) {

	bucket := strings.Split(strings.Split(url, ".")[0], "//")[1]
	object := strings.Split(url, ".com/")[1]
	localFileName := localDirectory + strings.Split(object, "/")[0] + "_" + strings.Split(object, "/")[2]

	err := downloadObjectToLocal(bucket, object, localFileName)
	if err != nil {
		return "", err
	}

	return strings.Split(localFileName, ".")[0], nil
}

func compressFile(ctx context.Context, filepath string, format model.Format) error {
	// Two pass encoding
	encodeCmdPass0, _ := createEncodeCommand(filepath, 1, format.Encode)
	argsPass0 := strings.Fields(encodeCmdPass0)
	cmdPass0 := exec.Command(argsPass0[0], argsPass0[1:]...)
	errPass0 := executeEncodeCommand(ctx, cmdPass0)
	if errPass0 != nil {
		return errPass0
	}

	encodeCmdPass1, _ := createEncodeCommand(filepath, 2, format.Encode)
	argsPass1 := strings.Fields(encodeCmdPass1)
	cmdPass1 := exec.Command(argsPass1[0], argsPass1[1:]...)
	errPass1 := executeEncodeCommand(ctx, cmdPass1)
	if errPass1 != nil {
		return errPass1
	}

	return nil
}

func createEncodeCommand(filepath string, pass int, encodes []model.Encode) (encodeCmd string, outputPath string) {
	encodeCmd = "time ffmpeg" + " -i " + filepath + ".mp4"

	for _, encode := range encodes {
		pixelFormat := "yuv420p"
		rate := strconv.Itoa(encode.BitRate) + "k"

		videoCodec, framerate, tagVideo := encode.VideoCodec, encode.FrameRate, "hvc1"
		bitRate, bufferSize, maxRate := rate, rate, rate
		preset, videoFormat := "medium", encode.VideoFormat

		outputPath := filepath + "_" + encode.VideoCodec + "_" + encode.Size

		encodeCmd +=
			" -pix_fmt " + pixelFormat +
				" -movflags " + "faststart" +
				" -vsync " + "1" +
				" -vcodec " + videoCodec +
				" -r " + strconv.Itoa(framerate) +
				" -threads " + os.Getenv("FFMPEG_THREAD_COUNT") +
				" -b:v: " + bitRate +
				" -bufsize " + bufferSize +
				" -maxrate " + maxRate +
				" -preset " + preset +
				" -pass " + strconv.Itoa(pass) +
				" -f " + videoFormat +
				" -passlogfile " + outputPath

		if videoCodec == "libx265" {
			encodeCmd += " -tag:v " + tagVideo
		}
		if pass == 1 {
			encodeCmd += " /dev/null -y"
		} else {
			encodeCmd += " -y " + outputPath
		}
	}

	return
}

func executeEncodeCommand(ctx context.Context, cmd *exec.Cmd) error {
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		logger := activity.GetLogger(ctx).With(zap.String("HostID", HostID))
		logger.Info("Error running the command" + ": " + stderr.String())
		return errors.New(stderr.String())
	}
	return nil
}

func uploadFile(ctx context.Context, fpath string, format model.Format) error {
	gsContext := context.Background()
	storageClient, err := storage.NewClient(gsContext,
		option.WithCredentialsJSON([]byte(os.Getenv("GOOGLE_JSON"))))
	if err != nil {
		return err
	}
	defer storageClient.Close()
	for _, encode := range format.Encode {
		// bucket := strings.Split(strings.Split(encode.Destination, ".")[0], "//")[1]
		// object := strings.Split(encode.Destination, ".com/")[1]
		pathArr := strings.Split(encode.Destination, "/")
		bucket := pathArr[3]
		object := strings.Split(encode.Destination, pathArr[3]+"/")[1]
		filepath := fpath + "_" + encode.VideoCodec + "_" + encode.Size
		file, err := os.Open(filepath)
		writeContext := storageClient.Bucket(bucket).Object(object).NewWriter(gsContext)
		writeContext.ACL = []storage.ACLRule{{Role: storage.RoleReader, Entity: storage.AllUsers}}
		writeContext.CacheControl = "public, max-age=" + os.Getenv("CHACHE_AGE")
		if _, err = io.Copy(writeContext, file); err != nil {
			return err
		}
		if err := writeContext.Close(); err != nil {
			return err
		}
		defer file.Close()
	}

	files, err := path.Glob(fmt.Sprintf("%v*", fpath))
	if err != nil {
		return err
	}
	for _, f := range files {
		if err := os.Remove(f); err != nil {
			panic(err)
		}
	}
	return nil
}

func migrateToColdline(ctx context.Context, jobID string, format model.Format) error {
	gsContext := context.Background()

	gsContext, cancel := context.WithTimeout(gsContext, time.Minute*10)
	defer cancel()

	storageClient, err := storage.NewClient(gsContext,
		option.WithCredentialsJSON([]byte(os.Getenv("GOOGLE_JSON"))))
	if err != nil {
		return nil
	}
	defer storageClient.Close()

	bucket := strings.Split(strings.Split(format.Source, ".")[0], "//")[1]
	gcsObject := strings.Split(format.Source, ".com/")[1]

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

	return nil
}

func downloadObjectToLocal(bucket, object, localDirectory string) error {
	ctx := context.Background()
	client, err := storage.NewClient(ctx,
		option.WithCredentialsJSON([]byte(os.Getenv("GOOGLE_JSON"))))
	if err != nil {
		return err
	}
	defer client.Close()

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
