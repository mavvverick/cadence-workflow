package jobprocessor

import (
	"image"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"os"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/fogleman/gg"
	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
)

type DrawPoster struct {
	BG   string
	Logo string
	User User
}

type User struct {
	Name  string
	Image string
	Font  string
}

func (d *DrawPoster) BuildImage() error {
	background, err := d.LoadImage(d.BG)
	if err != nil {
		return err
	}

	profile, err := d.LoadImage(d.User.Image)
	if err != nil {
		return err
	}

	profile = imaging.Resize(profile, 104*1.5, 104*1.5, imaging.Lanczos)

	xBG := (background.Bounds().Max.X - 1) / 2
	factor := 1.5
	x1Profile, y1Profile := factor*128, factor*176
	_, y2Profile := factor*232, factor*280

	mask := gg.NewContext(background.Bounds().Max.X, background.Bounds().Max.Y)

	mask.DrawCircle(float64(xBG), (y2Profile+y1Profile)/2, (y2Profile-y1Profile)/2)
	mask.SetRGB(0, 0, 0)
	mask.Fill()
	mask.SetMask(mask.AsMask())
	mask.DrawImage(profile, int(x1Profile), int(y1Profile))

	poster := gg.NewContextForImage(background)
	poster.SetRGB(1, 1, 1)
	if err := poster.LoadFontFace(d.User.Font, 16*1.5); err != nil {
		return err
	}

	_, _ = factor*110, factor*292
	_, y2Text := factor*251, factor*318
	ringSize := float64(5)

	poster.DrawStringAnchored(d.User.Name, float64(xBG), y2Text, 0.5, 0.5)
	poster.DrawCircle(float64(xBG), (y2Profile+y1Profile)/2, (y2Profile-y1Profile+ringSize)/2)
	poster.SetRGB(1, 1, 1)
	poster.Fill()
	poster.DrawImage(mask.Image(), 0, 0)

	err = poster.SavePNG("/tmp/resources/" + d.User.Name + ".png")
	if err != nil {
		return err
	}
	return nil
}

func (d *DrawPoster) LoadImage(filename string) (image.Image, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	if _, err := f.Seek(0, 0); err != nil {
		return nil, err
	}
	var img image.Image
	ext := strings.Split(filename, ".")[1]

	if ext == "jpg" {
		img, err = jpeg.Decode(f)
		if err != nil {
			return nil, err
		}
	} else if ext == "png" {
		img, err = png.Decode(f)
		if err != nil {
			return nil, err
		}
	}
	return img, nil
}

func (d *DrawPoster) LoadFont(filename string) (*truetype.Font, error) {
	fontBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	font, err := freetype.ParseFont(fontBytes)
	if err != nil {
		return nil, err
	}
	return font, nil
}

func (d *DrawPoster) SaveImage(filename string, img image.Image) (image.Image, error) {
	f, err := os.Create(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	err = png.Encode(f, img)
	if err != nil {
		return nil, err
	}
	return img, nil
}
