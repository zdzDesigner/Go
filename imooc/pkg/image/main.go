package main

import (
	"image"
	"image/color"
	"image/png"
	"os"
)

func main() {
	img := image.NewNRGBA(image.Rect(0, 0, 200, 200))
	blue := color.RGBA{R: 0, G: 0, B: 255, A: 255}

	for x := 0; x < 200; x++ {
		for y := 0; y < 200; y++ {
			img.Set(x, y, blue)
		}
	}

	blue_file, err := os.Create("blue.png")
	if err != nil {
		return
	}

	if err := png.Encode(blue_file, img); err != nil {
		return
	}
	blue_file.Close()
}
