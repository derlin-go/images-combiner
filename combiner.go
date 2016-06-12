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

// ResizeTo is a wrapper around resize.Resize, which resize INPLACE
// img to the given width while preserving its aspect ratio.
// note that if the image width equals the new width, no manipulation
// is made.
func ResizeTo(img *image.Image, width int) {
	if ((*img).Bounds().Max.X != width) {
		*img = resize.Resize(uint(width), 0, *img, resize.NearestNeighbor)
	}
}

// Open and decode an image from a file
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

// Try to interpret an hex string as an RGBA color. If the alpha
// component is missing, it will be interpreted as 0xFF (opaque)
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

	// --------- program arguments

	// register and parse program arguments
	flag.IntVar(&yGap, "gap", 0, "gap between images");
	flag.BoolVar(&opaque, "opaque", false, "replace transparency by white or bgColor (if defined)");
	flag.StringVar(&yGapStrColor, "gapColor", "", "color of gap between images, as an hex string");
	flag.StringVar(&bgStrColor, "bgColor", "", "replace alpha to (leave empty to keep transparency");
	flag.Parse();

	// handle the bgColor argument: convert str -> color + set the gap color to the bg color by default
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

	// if the gap color is defined, try to convert the string to a color
	if yGapStrColor != "" {
		// default to bgcolor
		var err error;
		yGapColor, err = ParseColor(yGapStrColor)
		if (err != nil) {
			fmt.Println(err)
			os.Exit(0)
		}
	}

	// --------- load and resize images

	image_paths := os.Args[len(os.Args) - flag.NArg():]
	images := make([]image.Image, len(image_paths))
	width := math.MaxInt32;

	// open and decode the images, while keeping track of the smallest width
	for idx, path := range (image_paths) {
		img, _, _ := OpenAndDecode(path)
		if img.Bounds().Max.X < width {
			width = img.Bounds().Max.X
		}
		images[idx] = img
	}

	// resize all the images
	height := 0
	for i, img := range (images) {
		ResizeTo(&img, width)
		height += img.Bounds().Max.Y + yGap
		images[i] = img
	}

	// -------- create the composite image

	// first, create a uniform image
	compositeImage := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(compositeImage, compositeImage.Bounds(), &image.Uniform{bgColor}, image.ZP, draw.Src)
	// the offset is where to draw the next image (default to 0,0)
	offset := image.ZP
	var gapImg *image.RGBA;

	// the gapImg is used/drawn only in case gap color != bg color
	if yGap > 0 && yGapColor != bgColor {
		gapImg = image.NewRGBA(image.Rect(0, 0, width, yGap))
		draw.Draw(gapImg, gapImg.Bounds(), &image.Uniform{yGapColor}, image.ZP, draw.Src)
	}

	// draw each image over the composite image
	for _, img := range (images) {
		draw.Draw(compositeImage, img.Bounds().Add(offset), img, image.ZP, draw.Over)
		offset.Y += img.Bounds().Max.Y
		if gapImg != nil { // gap color != bg color, draw it
			draw.Draw(compositeImage, img.Bounds().Add(offset), gapImg, image.ZP, draw.Over)
		}
		offset.Y += yGap
	}

	// Create a new file and write the composite image in png format
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