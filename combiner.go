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

// Create a struct to deal with pixel
type Pixel struct {
	Point image.Point
	Color color.Color
}

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

// Decode image.Image's pixel data into []*Pixel
func DecodePixelsFromImage(img image.Image, offsetX, offsetY int) []*Pixel {
	pixels := []*Pixel{}
	for y := 0; y <= img.Bounds().Max.Y; y++ {
		for x := 0; x <= img.Bounds().Max.X; x++ {
			p := &Pixel{
				Point: image.Point{x + offsetX, y + offsetY},
				Color: img.At(x, y),
			}
			pixels = append(pixels, p)
		}
	}
	return pixels
}

func CreateGap(x int, yGap int, offsetY int, color color.Color) []*Pixel {
	gapPixels := make([]*Pixel, x * yGap);
	idx := 0
	for i := offsetY; i < offsetY + yGap; i++ {
		for j := 0; j < x; j++ {
			gapPixels[idx] = &Pixel{
				Point: image.Point{j, i},
				Color: color,
			}
			idx++
		}
	}
	fmt.Println(len(gapPixels))
	return gapPixels
}

func ParseColor(hexString string) (color.Color, error) {
	bs, _ := hex.DecodeString(hexString)
	var alpha uint8 = 255;
	if 3 < len(bs) && len(bs) > 4 {
		return nil, errors.New("undefined color")
	}
	if len(bs) > 3 {
		alpha = bs[3]; }
	return color.RGBA{bs[0], bs[1], bs[2], alpha}, nil
}

func main() {

	var yGap int;
	var opaque bool;
	var yGapStrColor, bgStrColor string;
	var yGapColor color.Color;
	var bgColor color.Color = nil;

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

	if yGapStrColor != "" { // default to bgcolor
		var err error;
		yGapColor, err = ParseColor(yGapStrColor)
		if (err != nil) {
			fmt.Println(err)
			os.Exit(0)
		}
	}

	image_paths := os.Args[len(os.Args) - flag.NArg():];
	images := make([]image.Image, len(image_paths))
	var pixels []*Pixel;
	width := math.MaxInt32;

	for idx, path := range (image_paths) {
		img, _, _ := OpenAndDecode(path);
		if img.Bounds().Max.X < width {
			width = img.Bounds().Max.X
		}
		images[idx] = img
	}

	offset := 0;
	for i, img := range (images) {
		ResizeTo(&img, width)
		pixels = append(pixels, DecodePixelsFromImage(img, 0, offset)...);
		offset += img.Bounds().Max.Y;
		if yGap > 0 && i < len(images) - 1 {
			pixels = append(pixels, CreateGap(width, yGap, offset, yGapColor)...)
			offset += int(yGap);
		}
	}
	pixelSum := append(pixels)
	fmt.Print()
	// Set a new size for the new image equal to the max width
	// of bigger image and max height of two images combined
	newRect := image.Rectangle{
		Min: images[0].Bounds().Min,
		Max: image.Point{
			X: width,
			Y: offset,
		},
	}
	finImage := image.NewRGBA(newRect)
	// This is the cool part, all you have to do is loop through
	// each Pixel and set the image's color on the go
	for _, px := range pixelSum {
		finImage.Set(
			px.Point.X,
			px.Point.Y,
			px.Color,
		)
	}

	fmt.Println(finImage.Bounds())
	dst := image.NewRGBA(finImage.Bounds())
	fmt.Println(dst.Bounds())
	if bgColor != nil {
		draw.Draw(dst, dst.Bounds(), &image.Uniform{bgColor}, image.ZP, draw.Src)
		draw.Draw(dst, finImage.Bounds(), finImage, image.ZP, draw.Over)
	} else {
		draw.Draw(dst, finImage.Bounds(), finImage, image.ZP, draw.Src)
	}

	// Create a new file and write to it
	out, err := os.Create("./output.png")
	if err != nil {
		panic(err)
		os.Exit(1)
	}
	err = png.Encode(out, dst)
	if err != nil {
		panic(err)
		os.Exit(1)
	}
}