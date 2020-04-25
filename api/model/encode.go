package model

import (
	"context"
	jp "github.com/YOVO-LABS/workflow/workflows/jobprocessor"
)

// QueryParams ....
type QueryParams struct {
	Query *Query `json:"query"`
}

func (q QueryParams) Validate(ctx context.Context) error {
	return ValidateFields(q)
}

// Query ....
type Query struct {
	Source      string `json:"source"`
	Payload     string `json:"payload"`
	CallbackURL string `json:"callback_url"`
	Format      []*FormatBody
}

// FormatBody ...
type FormatBody struct {
	Output              string `json:"output"`
	Size                string `json:"size"`
	Bitrate             int    `json:"bitrate"`
	VideoCodec          string `json:"video_codec"`
	Destination         DestinationBody
	Framerate           int    `json:"framerate"`
	FileExtension       string `json:"file_extension"`
	VideoCodecParamters VideoCodecParamters
	TagVideo            string `json:"tag_video"`
	AudioBitrate        string `json:"audio_bitrate"`
	AudioSampleRate     string `json:"audio_sample_rate"`
	AudioChannelsNumber string `json:"audio_channels_number"`
	Logo 				jp.Logo
}

// DestinationBody ...
type DestinationBody struct {
	Key        string `json:"key"`
	Secret     string `json:"Secret"`
	URL        string `json:"url"`
	Permission string `json:"permission"`
}

// VideoCodecParamters ...
type VideoCodecParamters struct {
	VProfile   string  `json:"vprofile"`
	Level      int     `json:"level"`
	Coder      string  `json:"coder"`
	Flags2     string  `json:"flags2"`
	TwoPass    int     `json:"two_pass"`
	IQfactor   float32 `json:"i_qfactor"`
	Partitions string  `json:"partitions"`
	Directpred string  `json:"directpred"`
	MeMethod   string  `json:"me_method"`
	BStrategy  string  `json:"b_strategy"`
}