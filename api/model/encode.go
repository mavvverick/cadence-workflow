package model

import "context"

// Query ....
type Query struct {
	Source      string `json:"source"`
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

// Format ...
type Format struct {
	Source      string
	CallbackURL string
	Encode      []Encode
}

//BuildFormat ...
type BuildFormat interface {
	SetFormatSource(string) BuildFormat
	SetFormatCallbackURL(string) BuildFormat
	SetFormatEncode([]Encode) BuildFormat
	GetFormat() Format
}

//NewVideoFormat ...
func NewVideoFormat() VideoFormat {
	return VideoFormat{}
}

//VideoFormat ...
type VideoFormat struct {
	format Format
}

//SetFormatSource ...
func (v *VideoFormat) SetFormatSource(s string) BuildFormat {
	v.format.Source = s
	return v
}

//SetFormatCallbackURL ...
func (v *VideoFormat) SetFormatCallbackURL(s string) BuildFormat {
	v.format.CallbackURL = s
	return v
}

//SetFormatEncode ...
func (v *VideoFormat) SetFormatEncode(encodes []Encode) BuildFormat {
	for _, encode := range encodes {
		v.format.Encode = append(v.format.Encode, encode)
	}
	return v
}

//GetFormat ...
func (v *VideoFormat) GetFormat() Format {
	return v.format
}

// Validate ...
func (c *Query) Validate(ctx context.Context) error {
	return ValidateFields(c)
}

//Encode ...
type Encode struct {
	Source      string
	Destination string
	Size        string
	PixelFormat string
	VideoCodec  string
	FrameRate   int
	BitRate     int
	BufferSize  int
	MaxRate     int
	Preset      string
	VideoFormat string
}

//BuildEncode ...
type BuildEncode interface {
	SetSource(string) BuildEncode
	SetDestination(string) BuildEncode
	SetSize(string) BuildEncode
	SetPixelFormat(string) BuildEncode
	SetVideoCodec(string) BuildEncode
	SetFrameRate(int) BuildEncode
	SetBitRate(int) BuildEncode
	SetBufferSize(int) BuildEncode
	SetMaxRate(int) BuildEncode
	SetPreset(string) BuildEncode
	SetVideoFormat(string) BuildEncode
	GetEncode() Encode
}

// //Encoder ...
// type Encoder struct {
// 	builder BuildEncode
// }

// //SetBuilder ...
// func (e *Encoder) SetBuilder(b BuildEncode) {
// 	e.builder = b
// }

//NewMP4Encode ...
func NewMP4Encode() MP4Encode {
	return MP4Encode{}
}

//MP4Encode ...
type MP4Encode struct {
	encode Encode
}

//SetSource ...
func (l *MP4Encode) SetSource(s string) BuildEncode {
	l.encode.Source = s
	return l
}

//SetDestination ...
func (l *MP4Encode) SetDestination(s string) BuildEncode {
	l.encode.Destination = s
	return l
}

//SetSize ...
func (l *MP4Encode) SetSize(s string) BuildEncode {
	l.encode.Size = s
	return l
}

//SetPixelFormat ...
func (l *MP4Encode) SetPixelFormat(s string) BuildEncode {
	l.encode.PixelFormat = s
	return l
}

//SetVideoCodec ...
func (l *MP4Encode) SetVideoCodec(s string) BuildEncode {
	l.encode.VideoCodec = s
	return l
}

//SetFrameRate ...
func (l *MP4Encode) SetFrameRate(s int) BuildEncode {
	l.encode.FrameRate = s
	return l
}

//SetBitRate ...
func (l *MP4Encode) SetBitRate(s int) BuildEncode {
	l.encode.BitRate = s
	return l
}

//SetMaxRate ...
func (l *MP4Encode) SetMaxRate(s int) BuildEncode {
	l.encode.MaxRate = s
	return l
}

//SetPreset ...
func (l *MP4Encode) SetPreset(s string) BuildEncode {
	l.encode.Preset = s
	return l
}

//SetVideoFormat ...
func (l *MP4Encode) SetVideoFormat(s string) BuildEncode {
	l.encode.VideoFormat = s
	return l
}

//SetBufferSize ...
func (l *MP4Encode) SetBufferSize(s int) BuildEncode {
	l.encode.BufferSize = s
	return l
}

//GetEncode ...
func (l *MP4Encode) GetEncode() Encode {
	return l.encode
}
