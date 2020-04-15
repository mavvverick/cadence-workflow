package pkg

import (
	"fmt"
	"github.com/disintegration/gift"
	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"os"
	"strings"
)

type  DrawPoster struct {
	BG string
	Logo string
	User User
}

type User struct {
	Name string
	Image string
	Font string
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

	dst := image.NewNRGBA(background.Bounds())
	gift.New().Draw(dst, background)

	xBG := (background.Bounds().Max.X - 1) / 2

	xFGProfile := (profile.Bounds().Max.X - 1) / 2

	//xBG-xFGProfile returns the mean of leftmost boundary of background and profile
	x := xBG - xFGProfile
	yProfile := int(176 * 1.5) //As per the design
	gift.New().DrawAt(dst, profile, image.Pt(x, yProfile), gift.OverOperator)

	//constructing image using username
	title := d.User.Name

	//load the font type
	fontType, err := d.LoadFont(d.User.Font)
	if err != nil {
		return err
	}

	//Decide on text color, and text background
	tSrc, bg := image.White, image.Transparent

	//Set Width and Height for the rectangle that encpasulates text
	tWidth, tHeight := background.Bounds().Max.X, 26*1.5 //As per design
	tRect := image.Rect(0, 0, tWidth, int(tHeight))
	tRgba := image.NewRGBA(tRect)
	draw.Draw(tRgba, tRgba.Bounds(), bg, image.ZP, draw.Src)

	//Initialize a drawer with size and dpi
	drawer := &font.Drawer{
		Dst: tRgba,
		Src: tSrc,
		Face: truetype.NewFace(fontType, &truetype.Options{
			Size: 16 * 1.5,
			DPI:  72,
		}),
	}
	drawer.Dot = fixed.Point26_6{
		X: (fixed.I(tWidth) - drawer.MeasureString(title)) / 2,
		Y: fixed.I(int((tHeight + 10) / 2)),
	}

	drawer.DrawString(title)
	gift.New().DrawAt(dst, tRgba, image.Pt(0, int(292*1.5)), gift.OverOperator)

	dstPath := fmt.Sprintf("/tmp/%v.png", title)
	_, err = d.SaveImage(dstPath, dst)
	if err != nil {
		fmt.Println("error", err)
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