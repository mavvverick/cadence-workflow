package jobprocessor

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/YOVO-LABS/workflow/pkg"

	"github.com/YOVO-LABS/workflow/api/model"
	"google.golang.org/api/option"

	"cloud.google.com/go/storage"
	"go.uber.org/cadence/activity"
	"go.uber.org/zap"
)

/**
 * Sample activities used by file processing sample workflow.
 */
const (
	downloadFileActivityName  = "downloadFileActivity"
	compressMediaActivityName = "compressMediaActivity"
	addThumbnailActivityname  = "addThumbnailActivity"
	uploadFileActivityName    = "uploadFileActivity"

	migrateToColdLineActivityName = "migrateToColdLineActivity"
	srcDirectory                  = "raw/"
	processedDirectory            = "processed/"
	blackHole                     = "blackHole/"
	localDirectory                = "/tmp/"
	localProcessedDirectory       = "/tmp/"
	waterMarkFileName             = "watermark.gif"
	watermarkFolder               = "pilot"
)

// This is registration process where you register all your activity handlers.
func init() {
	activity.RegisterWithOptions(
		downloadFileActivity,
		activity.RegisterOptions{Name: downloadFileActivityName},
	)
	activity.RegisterWithOptions(
		compressMediaActivity,
		activity.RegisterOptions{Name: compressMediaActivityName},
	)
	activity.RegisterWithOptions(
		uploadFileActivity,
		activity.RegisterOptions{Name: uploadFileActivityName},
	)
}

func downloadFileActivity(ctx context.Context, jobID, url, payload, watermark string) (*model.DownloadObject, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Downloading file from gcs", zap.String("jobID", jobID))

	fmt.Println(jobID, time.Now(), "Download Activity -> Start")

	dO, err := downloadResources(ctx, url, payload, watermark)
	if err != nil {
		fmt.Println(jobID, time.Now(), DownloadActivityErrorMsg)
		return nil, err
	}

	fmt.Println(jobID, time.Now(), "Download Activity -> Finished")

	return dO, nil
}

func compressMediaActivity(ctx context.Context, jobID string, dO model.DownloadObject, format model.Format) error {
	logger := activity.GetLogger(ctx).With(zap.String("jobID", jobID))
	logger.Info("compressFileActivity started.", zap.String("FileName", dO.VideoPath))

	fmt.Println(jobID, time.Now(), "compressFile Activity -> Start")

	err := compressMedia(dO, format)
	if err != nil {
		fmt.Println(jobID, time.Now(), CompressionActivityErrorMsg)
		return err
	}

	fmt.Println(jobID, time.Now(), "compressFile Activity -> Finished")

	return nil
}

func uploadFileActivity(ctx context.Context, jobID, fpath string, format model.Format) error {
	logger := activity.GetLogger(ctx).With(zap.String("HostID", "ABCd"))
	logger.Info("uploadFileActivity begin", zap.String("FileName", localProcessedDirectory))

	fmt.Println(jobID, time.Now(), "uploadFile Activity -> Start")

	err := uploadFile(fpath, format)
	if err != nil {
		fmt.Println(jobID, time.Now(), UploadActivityErrorMsg)
		return err
	}

	fmt.Println(jobID, time.Now(), "uploadFile Activity -> Finished")
	return nil
}

// get video, watermark, user info required for compression
func downloadResources(ctx context.Context, url, payload, watermarkURL string) (*model.DownloadObject, error) {
	var dO model.DownloadObject

	bucket := strings.Split(strings.Split(url, "//")[1], ".")[0]
	objectPath := strings.Split(url, ".com/")[1]
	object := strings.Split(objectPath, "/")

	client, err := storage.NewClient(ctx,
		option.WithCredentialsJSON([]byte(os.Getenv("GOOGLE_JSON"))))
	if err != nil {
		return nil, err
	}
	defer client.Close()

	// download video to be encoded
	localFileName := localDirectory + object[0] + "_" + object[len(object)-1]
	dO.VideoPath = strings.Split(localFileName, ".")[0]
	err = downloadGCSObjectToLocal(ctx, client, bucket, objectPath, localFileName)
	if err != nil {
		return nil, err
	}

	//download watermark logo
	watermarkURLSplit := strings.Split(watermarkURL, "/")
	watermarkFileName := watermarkURLSplit[len(watermarkURLSplit)-1]
	if _, err := os.Stat(localDirectory + watermarkFileName); err != nil {
		err = downloadFileWithURL(localDirectory+watermarkFileName, watermarkURL)
		if err != nil {
			return nil, err
		}
	}
	dO.Watermark = localDirectory + watermarkFileName

	return &dO, nil
}

func compressMedia(dO model.DownloadObject, format model.Format) error {
	encodeCmd264, _, watermarkCmd, _ := createEncodeCommand(dO, format.Encode)

	//encode 240p video using x264
	err := executeCommand(encodeCmd264)
	if err != nil {
		return err
	}

	//encode 540p video using x265
	//err = executeCommand(encodeCmd265)
	//if err != nil {
	//	return err
	//}

	//add watermark to the earlier encoded 540p videos
	for _, wc := range watermarkCmd {
		fmt.Println(wc)
		err = executeCommand(wc)
		if err != nil {
			return err
		}
	}

	if err := os.Remove(dO.VideoPath + ".mp4"); err != nil {
		return err
	}
	return nil
}

func createEncodeCommand(dO model.DownloadObject, encodes []model.Encode) (encodeCmd264, encodeCmd265 string, watermarkCmd, thumbnailCmd []string) {
	encodeCmd264 = "ffmpeg" + " -i " + dO.VideoPath + ".mp4"
	encodeCmd265 = "ffmpeg" + " -i " + dO.VideoPath + ".mp4"

	for _, encode := range encodes {
		destFields := strings.Split(encode.Destination, "/")
		path := destFields[len(destFields)-2]
		outputPath := dO.VideoPath + "_" + encode.VideoCodec + "_" + encode.Size + "_" + path + ".mp4"

		if encode.VideoCodec == "libx265" {
			//encodeCmd265 += x265EncodeCmd(encode, outputPath)
		} else if encode.VideoCodec == "libx264" {
			encodeCmd264 += x264EncodeCmd(encode, outputPath)
		}
		if (model.Logo{}) != encode.Logo {
			watermarkCmd = append(watermarkCmd, createWatermarkCmd(encode, dO, "superfast"))
			//tc, err := createThumbnailCmd(dO, encode.VideoCodec, encode.Size)
			//if err == nil {
			//	thumbnailCmd = append(thumbnailCmd, tc)
			//}
		}
	}
	return
}

func createThumbnail(dO model.DownloadObject) error {
	imgPath := localDirectory + dO.UserImage
	poster := pkg.DrawPoster{
		BG: dO.Background,
		User: pkg.User{
			Name:  "@" + dO.UserImage,
			Image: imgPath + ".jpg",
			Font:  dO.Font,
		},
	}

	err := poster.BuildImage()
	if err != nil {
		return err
	}

	fileIn := imgPath + ".png"
	fileOut := imgPath + ".mp4"
	err = pngToMp4(fileIn, fileOut)
	if err != nil {
		return err
	}

	err = os.Remove(imgPath + ".jpg")
	if err != nil {
		return err
	}

	err = os.Remove(fileIn)
	if err != nil {
		return err
	}

	return nil
}

func uploadFile(fpath string, format model.Format) error {
	ctx := context.Background()
	storageClient, err := storage.NewClient(ctx,
		option.WithCredentialsJSON([]byte(os.Getenv("GOOGLE_JSON"))))
	if err != nil {
		return err
	}
	defer storageClient.Close()

	for _, encode := range format.Encode {
		//s3://storage.googleapis.com/yovo-test/test/T3/yo540p/1.mp4
		pathArr := strings.Split(encode.Destination, "/")
		bucket := pathArr[3]
		object := strings.Split(encode.Destination, pathArr[3]+"/")[1]

		filepath := fpath + "_" + encode.VideoCodec + "_" + encode.Size + "_" + pathArr[len(pathArr)-2] + ".mp4"
		err := uploadToGCS(ctx, *storageClient, filepath, bucket, object)
		if err != nil {
			return err
		}
		if (model.Logo{}) != encode.Logo {
			filepath = fpath + "_" + encode.VideoCodec + "_" + encode.Size + ".mp4"
			gcsPathField := strings.Split(object, "/"+pathArr[len(pathArr)-2])
			object = gcsPathField[0] + gcsPathField[1]
			err := uploadToGCS(ctx, *storageClient, filepath, bucket, object)
			if err != nil {
				return err
			}
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
	return nil
}
