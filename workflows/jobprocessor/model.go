package jobprocessor

// Format ...
type Format struct {
	Source      string
	CallbackURL string
	Payload     string
	Encode      []Encode
	WatermarkURL string
}

//Encode ...
type Encode struct {
	Source      	string
	Destination 	string
	Size        	string
	PixelFormat 	string
	VideoCodec  	string
	FrameRate   	int
	BitRate     	int
	BufferSize  	int
	MaxRate     	int
	Preset      	string
	VideoFormat 	string
	Logo 			Logo
}

type Logo struct {
	Source string
	X string
	Y string
}

type DownloadObject struct {
	VideoPath	string
	Watermark 	string
	UserImage 	string
	Background	string
	Font		string
	Meta        *Meta
}

type Meta struct {
	Type     string
	Duration float64
	Size     float64
	Bitrate  int
}