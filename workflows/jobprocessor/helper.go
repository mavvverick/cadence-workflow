package jobprocessor

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/YOVO-LABS/workflow/api/model"
)

var overlayWatermark = "[1:v]scale=%v:%v[v1];[0:v][v1]overlay='mod(trunc((t+0)/5),2)*(W-w-0)':" +
	"'mod(trunc((t+0)/5),2)*(H-h-0)':enable='gt(t,0)'%v"

var overlayThumbnail = "color=black:%v:d=%v[base];[0:v]setpts=PTS-STARTPTS[v0];" +
	"[1:v]format=yuva420p,fade=in:st=0:d=0.4:alpha=1,setpts=PTS-STARTPTS+((%v)/TB)[v1];" +
	"[base][v0]overlay[tmp];" +
	"[tmp][v2]overlay,format=yuv420p[fv];" +
	"[0:a]afade=out:st=%v:d=1"

func createWatermarkCmd(encode model.Encode, dO model.DownloadObject, preset string) string {
	destFields := strings.Split(encode.Destination, "/")
	path := destFields[len(destFields)-2]

	inputPath :=
		dO.VideoPath + "_" +
			encode.VideoCodec + "_" +
			encode.Size + "_" +
			path + ".mp4"

	watermarkCmd := "ffmpeg" + " -i " + inputPath
	filterComplex := ""

	extension := strings.Split(dO.Watermark, ".")[1]
	if extension == "gif" {
		watermarkCmd +=
			" -ignore_loop " + "0" +
				" -i " + dO.Watermark
		filterComplex = fmt.Sprintf(overlayWatermark, 194, 124, ":shortest=1")
	} else if extension == "png" {
		watermarkCmd += " -i " + dO.Watermark
		filterComplex = fmt.Sprintf(overlayWatermark, 146, 60, "")
	}

	outputPath :=
		dO.VideoPath + "_" +
			encode.VideoCodec + "_" +
			encode.Size + ".mp4"

	watermarkCmd +=
		" -filter_complex " + filterComplex +
			" -vcodec " + encode.VideoCodec +
			" -preset " + preset +
			" -y " + outputPath
	return watermarkCmd
}

func createThumbnailCmd(dO model.DownloadObject, codec, size string) (string, error) {
	duration, err := getMediaDuration(dO.VideoPath + ".mp4")
	if err != nil {
		return "", nil
	}

	inputFilePath := dO.VideoPath + "_" + codec + "_" + size + ".mp4"
	watermarkFilePath := dO.VideoPath + "_" + codec + "_" + size + "_" + "wm" + ".mp4"

	filter := fmt.Sprintf(overlayThumbnail, size, duration+3-0.4, duration-0.4, duration-1.0)

	thumbnailCmd :=
		"ffmpeg" +
			" -i " + watermarkFilePath +
			" -i " + localDirectory + dO.UserImage + ".png" +
			" -filter_complex " + filter +
			" -preset " + "superfast" +
			" -map " + "[fv]" +
			" -y " + inputFilePath

	// remove the tmp file
	err = os.Remove(watermarkFilePath)
	if err != nil {
		return "", nil
	}
	return thumbnailCmd, nil
}

func getMediaDuration(fpath string) (float64, error) {
	durationCmd := "ffprobe -i " + fpath + " -show_format -v quiet"
	duration, err := executeCommandWithOutput(durationCmd)
	if err != nil {
		return -1, err
	}

	fmt.Println(strings.Split(duration, "\n"))
	duration = strings.Split(strings.Split(duration, "\n")[7], "=")[1]
	durationInt, err := strconv.ParseFloat(duration, 8)
	if err != nil {
		return -1, nil
	}
	return durationInt, nil
}

func pngToMp4(ImgPath, mp4Path string) error {
	cmd := "ffmpeg" +
		" -loop " + "1" +
		" -i " + ImgPath +
		" -c:v " + "libx264" +
		" -t " + "3" +
		" -pix_fmt " + "yuv420p" +
		" -profile:v " + "high" +
		" -crf " + "20" +
		" -preset " + "ultrafast" +
		" -y " + mp4Path

	err := executeCommand(cmd)
	if err != nil {
		return err
	}
	return nil
}

func executeCommand(execCmd string) error {
	var stderr bytes.Buffer

	args := strings.Fields(execCmd)
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return errors.New(stderr.String())
	}
	return nil
}

func executeCommandWithOutput(execCmd string) (string, error) {
	var stderr, stdout bytes.Buffer

	args := strings.Fields(execCmd)
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stderr = &stderr
	cmd.Stdout = &stderr

	err := cmd.Run()
	if err != nil {
		return "", errors.New(stderr.String())
	}
	fmt.Println(stdout.String(), stderr.String())
	outStr := stderr.String()
	return outStr, nil
}

func x265EncodeCmd(encode model.Encode, outputPath string) string {
	bitrate := strconv.Itoa(encode.BitRate) + "k"
	x265Cmd :=
		" -pix_fmt " + "yuv420p" +
			" -movflags " + "faststart" +
			" -vcodec " + encode.VideoCodec +
			" -r " + strconv.Itoa(encode.FrameRate) +
			" -threads " + os.Getenv("FFMPEG_THREAD_COUNT") +
			" -b:v: " + bitrate +
			" -bufsize " + bitrate +
			" -maxrate " + bitrate +
			" -preset " + "superfast" + // faster execution time at the expense of quality
			" -s " + encode.Size +
			" -profile:v " + "high" +
			" -level " + "4.0" +
			" -f " + encode.VideoFormat +
			//" -tag:v " 		+ "hvc1" +
			" -y " + outputPath
	return x265Cmd
}

func x264EncodeCmd(encode model.Encode, outputPath string) string {
	bitrate := strconv.Itoa(encode.BitRate) + "k"
	x264Cmd :=
		" -pix_fmt " + "yuv420p" +
			" -movflags " + "faststart" +
			" -vcodec " + encode.VideoCodec +
			" -r " + strconv.Itoa(encode.FrameRate) +
			" -threads " + os.Getenv("FFMPEG_THREAD_COUNT") +
			" -b:v: " + bitrate +
			" -bufsize " + bitrate +
			" -maxrate " + bitrate +
			" -preset " + "superfast" +
			" -s " + encode.Size +
			" -profile:v " + "high" +
			" -level " + "4.0" +
			" -f " + encode.VideoFormat +
			" -y " + outputPath
	return x264Cmd
}

func downloadGCSObjectToLocal(ctx context.Context, client *storage.Client, bucket, object, localDirectory string) error {

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

func downloadFileWithURL(filepath string, url string) error {
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return errors.New("not found")
	}

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

func uploadToGCS(ctx context.Context, sc storage.Client, filepath, bucket, object string) error {
	f, err := os.Open(filepath)
	if err != nil {
		return err
	}
	writeContext := sc.Bucket(bucket).Object(object).NewWriter(ctx)
	writeContext.ACL = []storage.ACLRule{{Role: storage.RoleReader, Entity: storage.AllUsers}}
	writeContext.CacheControl = "public, max-age=" + os.Getenv("CHACHE_AGE")

	if _, err = io.Copy(writeContext, f); err != nil {
		return err
	}
	if err := writeContext.Close(); err != nil {
		return err
	}
	f.Close()
	if err := os.Remove(filepath); err != nil {
		return err
	}
	return nil
}
