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
)

// Create a struct to deal with pixel
type Pixel struct {
	Point image.Point
	Color color.Color
}

func ResizeTo(img *image.Image, x int) {
	if ((*img).Bounds().Max.X != x) {
		*img = resize.Resize(uint(x), 0, *img, resize.Lanczos3)
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

func main() {

	image_paths := os.Args[1:];

	images := make([]image.Image, len(image_paths))
	var pixels []*Pixel;
	x := math.MaxInt32;

	for idx, path := range (image_paths) {
		img, _, _ := OpenAndDecode(path);
		if img.Bounds().Max.X < x {
			x = img.Bounds().Max.X
		}
		images[idx] = img
	}

	offset := 0;
	for _, img := range (images) {
		ResizeTo(&img, x)
		pixels = append(pixels, DecodePixelsFromImage(img, 0, offset)...);
		offset += img.Bounds().Max.Y;
	}
	pixelSum := append(pixels)
	fmt.Print()
	// Set a new size for the new image equal to the max width
	// of bigger image and max height of two images combined
	newRect := image.Rectangle{
		Min: images[0].Bounds().Min,
		Max: image.Point{
			X: x,
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
	draw.Draw(finImage, finImage.Bounds(), finImage, image.Point{0, 0}, draw.Src)

	// Create a new file and write to it
	out, err := os.Create("./output.png")
	if err != nil {
		panic(err)
		os.Exit(1)
	}
	err = png.Encode(out, finImage)
	if err != nil {
		panic(err)
		os.Exit(1)
	}
}