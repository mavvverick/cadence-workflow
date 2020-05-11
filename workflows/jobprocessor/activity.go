package jobprocessor

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/uber/cadence/common"
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
	localTmpDirectory             = "/tmp/"
	waterMarkFileName             = "watermark.gif"
	watermarkFolder               = "pilot"
)

var localDirectory = common.StringPtr("/tmp/")

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

func downloadFileActivity(ctx context.Context, jobID, url, payload, watermark string, cb *CallbackInfo) (*DownloadObject, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Downloading file from gcs", zap.String("jobID", jobID))

	fmt.Println(jobID, time.Now(), "Download Activity -> Start")

	dO, err := downloadResources(ctx, url, payload, watermark)
	if err != nil {
		fmt.Println(jobID, time.Now(), DownloadActivityErrorMsg)
		cb.PushMessage(ctx, Download, Task, jobID, CallbackErrorEvent)
		return nil, err
	}

	fmt.Println(jobID, time.Now(), "Download Activity -> Finished")

	return dO, nil
}

func compressMediaActivity(ctx context.Context, jobID string, dO DownloadObject, format Format, cb *CallbackInfo) error {
	logger := activity.GetLogger(ctx)
	logger.Info("compressFileActivity started.", zap.String("jobID", jobID))

	fmt.Println(jobID, time.Now(), "compressFile Activity -> Start")

	err := compressMedia(dO, format)
	if err != nil {
		fmt.Println(jobID, time.Now(), CompressionActivityErrorMsg)
		cb.PushMessage(ctx, Compression, Task, jobID, CallbackErrorEvent)
		return err
	}

	fmt.Println(jobID, time.Now(), "compressFile Activity -> Finished")

	return nil
}

func uploadFileActivity(ctx context.Context, jobID, fpath string, format Format, cb *CallbackInfo) error {
	logger := activity.GetLogger(ctx)
	logger.Info("uploadFileActivity begin", zap.String("jobID", jobID))

	fmt.Println(jobID, time.Now(), "uploadFile Activity -> Start")

	err := uploadFile(fpath, format)
	if err != nil {
		fmt.Println(jobID, time.Now(), UploadActivityErrorMsg)
		cb.PushMessage(ctx, Upload, Task, jobID, CallbackErrorEvent)
		return err
	}

	fmt.Println(jobID, time.Now(), "uploadFile Activity -> Finished")
	cb.PushMessage(ctx, Completed, Task, jobID, "saved")
	return nil
}

// get video, watermark, user info required for compression
func downloadResources(ctx context.Context, url, payload, watermarkURL string) (*DownloadObject, error) {
	var dO DownloadObject

	bucket := strings.Split(strings.Split(url, "//")[1], ".")[0]
	objectPath := strings.Split(url, ".com/")[1]
	object := strings.Split(objectPath, "/")

	client, err := storage.NewClient(ctx,
		option.WithCredentialsJSON([]byte(os.Getenv("GOOGLE_JSON"))))
	if err != nil {
		return nil, err
	}
	defer client.Close()

	err = os.MkdirAll(*localDirectory+"resources/", 0770)
	if err != nil {
		return nil, err
	}

	localDirectory = common.StringPtr(localTmpDirectory+"resources/")

	// download video to be encoded
	localFileName := *localDirectory + object[0] + "_" + object[len(object)-1]
	dO.VideoPath = strings.Split(localFileName, ".")[0]
	err = downloadGCSObjectToLocal(ctx, client, bucket, objectPath, localFileName)
	if err != nil {
		return nil, err
	}

	//download watermark logo
	watermarkURLSplit := strings.Split(watermarkURL, "/")
	watermarkFileName := watermarkURLSplit[len(watermarkURLSplit)-1]
	if _, err := os.Stat(localTmpDirectory + watermarkFileName); err != nil {
		err = downloadFileWithURL(localTmpDirectory+watermarkFileName, watermarkURL)
		if err != nil {
			return nil, err
		}
	}
	dO.Watermark = localTmpDirectory + watermarkFileName

	//download background and font for poster
	payloadFields := strings.Split(payload, "|")
	if len(payloadFields) == 0 || len(payloadFields) < 3 {
		return &dO, nil
	} else {
		bucket = "yovo-app"
		objectPath = "pro_images/" + payloadFields[2] + ".jpg"
		userFileName := *localDirectory + payloadFields[2] + ".jpg"

		//download user profile photo
		err = downloadGCSObjectToLocal(ctx, client, bucket, objectPath, userFileName)
		if err ==  storage.ErrObjectNotExist {
			objectPath := "assets/background2.mp4"
			thumbnailBG := localTmpDirectory + "bg2.mp4"
			if _, err := os.Stat(thumbnailBG); err != nil {
				//user image doesn't exist, download the alternate thumbnail video
				err = downloadGCSObjectToLocal(ctx, client, bucket, objectPath, thumbnailBG)
				if err != nil {
					return nil, err
				}
			}
			os.Link(thumbnailBG, *localDirectory + payloadFields[2] + ".mp4")
			dO.UserImage = payloadFields[2]
			return &dO, nil
		} else if err != nil {
			return nil, err
		}
		dO.UserImage = payloadFields[2]

		objectPath = "assets/background.png"
		thumbnailBG := localTmpDirectory + "bg.png"
		if _, err := os.Stat(thumbnailBG); err != nil {
			//download background for the thumbnail
			err = downloadGCSObjectToLocal(ctx, client, bucket, objectPath, thumbnailBG)
			if err ==  storage.ErrObjectNotExist {
				return &dO, nil
			} else if err != nil {
				return nil, err
			}
		}
		dO.Background = thumbnailBG

		objectPath = "assets/font.ttf"
		font := localTmpDirectory + "font.ttf"
		if _, err := os.Stat(font); err != nil {
			//download font used for the thumbnail
			err = downloadGCSObjectToLocal(ctx, client, bucket, objectPath, font)
			if err ==  storage.ErrObjectNotExist {
				return &dO, err
			} else if err != nil {
				return nil, err
			}
		}
		dO.Font = font
	}
	return &dO, nil
}

func compressMedia(dO DownloadObject, format Format) error {
	_, err := getMediaMeta(&dO)
	if err != nil {
		return err
	}

	encodeCmd264, _, watermarkCmd, thumbnailCmd := createEncodeCommand(dO, format.Encode)
	//encode 240p video using x264
	err = executeCommand(encodeCmd264)
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
		err = executeCommand(wc)
		if err != nil {
			return err
		}
	}

	err = createThumbnail(dO)
	if err != nil {
		return err
	} else {
		for _, tc := range thumbnailCmd {
			err = executeCommand(tc)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func createEncodeCommand(dO DownloadObject, encodes []Encode) (encodeCmd264, encodeCmd265 string, watermarkCmd, thumbnailCmd []string) {
	encodeCmd264 = "ffmpeg" + " -i " + dO.VideoPath + ".mp4"
	encodeCmd265 = "ffmpeg" + " -i " + dO.VideoPath + ".mp4"

	for _, encode := range encodes {
		destFields := strings.Split(encode.Destination, "/")
		path := destFields[len(destFields)-2]
		outputPath := dO.VideoPath + "_" + encode.VideoCodec + "_" + encode.Size + "_" + path + ".mp4"
		if encode.BitRate > dO.Meta.Bitrate {
			encode.BitRate = dO.Meta.Bitrate
		}
		if encode.VideoCodec == "libx265" {
			//encodeCmd265 += x265EncodeCmd(encode, outputPath)
		} else if encode.VideoCodec == "libx264" {
			encodeCmd264 += x264EncodeCmd(encode, outputPath)
		}
		if (Logo{}) != encode.Logo {
			watermarkCmd = append(watermarkCmd, createWatermarkCmd(encode, dO, "veryfast"))
			tc, err := createThumbnailCmd(dO, encode.VideoCodec, encode.Size)
			if err == nil {
				thumbnailCmd = append(thumbnailCmd, tc)
			}
		}
	}
	return
}

func createThumbnail(dO DownloadObject) error {
	imgPath := *localDirectory + dO.UserImage

	if dO.Background != "" {
		poster := DrawPoster{
			BG: dO.Background,
			User: User{
				Name:  "@" + dO.UserImage,
				Image: imgPath + ".jpg",
				Font:  dO.Font,
			},
		}
		err := poster.BuildImage()
		if err != nil {
			return err
		}
		err = pngToMp4(*localDirectory+poster.User.Name+".png", *localDirectory + dO.UserImage+".mp4")
		if err != nil {
			return err
		}
	}
	return nil
}

func uploadFile(fpath string, format Format) error {
	ctx := context.Background()
	storageClient, err := storage.NewClient(ctx,
		option.WithCredentialsJSON([]byte(os.Getenv("GOOGLE_JSON"))))
	if err != nil {
		return err
	}
	defer storageClient.Close()

	for _, encode := range format.Encode {
		pathArr := strings.Split(encode.Destination, "/")
		bucket := pathArr[3]
		object := strings.Split(encode.Destination, pathArr[3]+"/")[1]

		filepath := fpath + "_" + encode.VideoCodec + "_" + encode.Size + "_" + pathArr[len(pathArr)-2] + ".mp4"
		err := uploadToGCS(ctx, *storageClient, filepath, bucket, object)
		if err != nil {
			return err
		}
		if (Logo{}) != encode.Logo {
			filepath = fpath + "_" + encode.VideoCodec + "_" + encode.Size + ".mp4"
			gcsPathField := strings.Split(object, "/"+pathArr[len(pathArr)-2])
			object = gcsPathField[0] + gcsPathField[1]
			err := uploadToGCS(ctx, *storageClient, filepath, bucket, object)
			if err != nil {
				return err
			}
		}
	}

	if err = os.RemoveAll(*localDirectory); err != nil {
		return err
	}
	return nil
}

func migrateToColdline(ctx context.Context, jobID string, format Format) error {
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
