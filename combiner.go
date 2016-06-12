package combiner

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
    "encoding/hex"
    "errors"
    "fmt"
    "bytes"
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

func DefaultCompose(images []*image.Image) ([]byte, error) {
    return Compose(images, color.White, false, 0, nil)
}

func Compose(images []*image.Image, bgColor color.Color, opaque bool, yGap int, yGapColor color.Color) ([]byte, error){

    // find min width
    width := math.MaxInt32
    for i, img := range (images) {
        fmt.Printf("   IMAGE %d : %s", i, (*img).Bounds())
        if (*img).Bounds().Max.X < width {
            width = (*img).Bounds().Max.X
        }
    }

    // resize all the images
    height := 0
    for i, img := range (images) {
        ResizeTo(img, width)
        height += (*img).Bounds().Max.Y + yGap
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
        draw.Draw(compositeImage, (*img).Bounds().Add(offset), (*img), image.ZP, draw.Over)
        offset.Y += (*img).Bounds().Max.Y
        if gapImg != nil {
            // gap color != bg color, draw it
            draw.Draw(compositeImage, (*img).Bounds().Add(offset), gapImg, image.ZP, draw.Over)
        }
        offset.Y += yGap
    }

    // Create a new file and write the composite image in png format

    buf := new(bytes.Buffer)
    err := png.Encode(buf, compositeImage)
    if err != nil {
        return nil, err
    }

    return buf.Bytes(), nil
}
