package main

import (
	"image"
	"image/color"
	"os"
	"image/draw"
	_ "image/png"
	_ "image/jpeg"
	"image/png"
	"github.com/nfnt/resize"
	"math"
	"fmt"
	"flag"
	"encoding/hex"
	"errors"
)

func ResizeTo(img *image.Image, x int) {
	if ((*img).Bounds().Max.X != x) {
		*img = resize.Resize(uint(x), 0, *img, resize.NearestNeighbor)
	}
}

// Keep it DRY so don't have to repeat opening file and decode
func OpenAndDecode(filepath string) (image.Image, string, error) {
	imgFile, err := os.Open(filepath)
	if err != nil {
		panic(err)
	}
	defer imgFile.Close()
	img, format, err := image.Decode(imgFile)
	if err != nil {
		panic(err)
	}
	return img, format, nil
}

func ParseColor(hexString string) (color.Color, error) {
	bs, _ := hex.DecodeString(hexString)
	var alpha uint8 = 255;
	if 3 < len(bs) && len(bs) > 4 {
		return nil, errors.New("undefined color")
	}
	if len(bs) > 3 {
		alpha = bs[3];
	}
	return color.RGBA{bs[0], bs[1], bs[2], alpha}, nil
}

func main() {

	var yGap int;
	var opaque bool;
	var yGapStrColor, bgStrColor string;
	var yGapColor color.Color;
	var bgColor color.Color = color.Transparent;

	flag.IntVar(&yGap, "gap", 0, "gap between images");
	flag.BoolVar(&opaque, "opaque", false, "replace transparency by white or bgColor (if defined)");
	flag.StringVar(&yGapStrColor, "gapColor", "", "color of gap between images, as an hex string");
	flag.StringVar(&bgStrColor, "bgColor", "", "replace alpha to (leave empty to keep transparency");
	flag.Parse();


	if opaque || bgStrColor != "" {
		if bgStrColor != "" {
			var err error;
			bgColor, err = ParseColor(bgStrColor)
			if (err != nil) {
				fmt.Println(err)
				os.Exit(0)
			}
		} else {
			bgColor = color.White;
		}
		yGapColor = bgColor;
	}

	if yGapStrColor != "" {
		// default to bgcolor
		var err error;
		yGapColor, err = ParseColor(yGapStrColor)
		if (err != nil) {
			fmt.Println(err)
			os.Exit(0)
		}
	}

	image_paths := os.Args[len(os.Args) - flag.NArg():]
	images := make([]image.Image, len(image_paths))
	width := math.MaxInt32;

	for idx, path := range (image_paths) {
		img, _, _ := OpenAndDecode(path)
		if img.Bounds().Max.X < width {
			width = img.Bounds().Max.X
		}
		images[idx] = img
	}

	height := 0
	for i, img := range (images) {
		ResizeTo(&img, width)
		height += img.Bounds().Max.Y + yGap
		images[i] = img
	}

	compositeImage := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(compositeImage, compositeImage.Bounds(), &image.Uniform{bgColor}, image.ZP, draw.Src)
	offset := image.ZP
	var gapImg *image.RGBA;

	if yGap > 0 && yGapColor != bgColor {
		gapImg = image.NewRGBA(image.Rect(0, 0, width, yGap))
		draw.Draw(gapImg, gapImg.Bounds(), &image.Uniform{yGapColor}, image.ZP, draw.Src)
	}

	for _, img := range (images) {
		draw.Draw(compositeImage, img.Bounds().Add(offset), img, image.ZP, draw.Over)
		offset.Y += img.Bounds().Max.Y
		if gapImg != nil {
			draw.Draw(compositeImage, img.Bounds().Add(offset), gapImg, image.ZP, draw.Over)
		}
		offset.Y += yGap
	}

	// Create a new file and write to it
	out, err := os.Create("./output.png")
	if err != nil {
		panic(err)
		os.Exit(1)
	}
	err = png.Encode(out, compositeImage)
	if err != nil {
		panic(err)
		os.Exit(1)
	}
}