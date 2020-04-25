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
)

var overlayWatermark = "[1:v]scale=%v:%v[v1];[0:v][v1]overlay='mod(trunc((t+0)/5),2)*(W-w-0)':" +
	"'mod(trunc((t+0)/5),2)*(H-h-0)':enable='gt(t,0)'%v"

var overlayThumbnail = "color=black:%v:d=%v[base];[0:v]setpts=PTS-STARTPTS[v0];" +
	"[1:v]format=yuva420p,fade=in:st=0:d=0.4:alpha=1,setpts=PTS-STARTPTS+((%v)/TB)[v1];" +
	"[base][v0]overlay[tmp];" +
	"[tmp][v1]overlay,format=yuv420p[fv];" +
	"[0:a]afade=out:st=%v:d=1"

func createWatermarkCmd(encode Encode, dO DownloadObject, preset string) string {
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
			encode.Size
	if dO.UserImage != "" {
		outputPath += "_wm.mp4"
	} else {
		outputPath += ".mp4"
	}

	watermarkCmd +=
		" -filter_complex " + filterComplex +
			" -vcodec " + encode.VideoCodec +
			" -preset " + preset +
			" -y " + outputPath
	return watermarkCmd
}

func createThumbnailCmd(dO DownloadObject, codec, size string) (string, error) {
	duration := dO.Meta.Duration
	inputFilePath := dO.VideoPath + "_" + codec + "_" + size + ".mp4"
	watermarkFilePath := dO.VideoPath + "_" + codec + "_" + size + "_" + "wm" + ".mp4"

	filter := fmt.Sprintf(overlayThumbnail, size, duration+3-0.4, duration-0.4, duration-1.0)

	thumbnailCmd :=
		"ffmpeg" +
			" -i " + watermarkFilePath +
			" -i " + *localDirectory + dO.UserImage + ".mp4" +
			" -filter_complex " + filter +
			" -preset " + "superfast" +
			" -map " + "[fv]" +
			" -y " + inputFilePath

	return thumbnailCmd, nil
}

func getMediaMeta(dO *DownloadObject) (*Meta, error) {
	probeCmd := "ffprobe -i " + dO.VideoPath + ".mp4" + " -show_format -v quiet"
	probe, err := executeCommandWithOutput(probeCmd)
	if err != nil {
		return nil, err
	}
	meta := strings.Split(probe, "\n")

	duration := strings.Split(meta[7], "=")[1]
	durationParsed, err := strconv.ParseFloat(duration, 8)
	if err != nil {
		return nil, err
	}

	size := strings.Split(meta[8], "=")[1]
	sizeParsed, err := strconv.ParseFloat(size, 16)
	if err != nil {
		return nil, err
	}

	bitrate := strings.Split(meta[9], "=")[1]
	bitrateParsed, err := strconv.ParseFloat(bitrate, 16)
	if err != nil {
		return nil, err
	}

	dO.Meta = &Meta{
		Duration: durationParsed,
		Size:     sizeParsed/(1024*1024),
		Bitrate:  int(bitrateParsed/1000),
	}

	return dO.Meta, nil
}

func pngToMp4(ImgPath, mp4Path string) error {
	cmd := "ffmpeg" +
		" -loop " + "1" +
		" -i " + ImgPath +
		" -c:v " + "libx264" +
		" -t " + "3" +
		" -crf " + "1" +
		" -preset " + "ultrafast" +
		" -pix_fmt " + "yuv420p" +
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
	var stderr, _ bytes.Buffer

	args := strings.Fields(execCmd)
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stderr = &stderr
	cmd.Stdout = &stderr

	err := cmd.Run()
	if err != nil {
		return "", errors.New(stderr.String())
	}
	outStr := stderr.String()
	return outStr, nil
}

func x265EncodeCmd(encode Encode, outputPath string) string {
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

func x264EncodeCmd(encode Encode, outputPath string) string {
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
			" -crf " + "31" +
			" -acodec " + "copy" +
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
	return nil
}
