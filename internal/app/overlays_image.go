package app

import (
	"bytes"
	"image"
	"image/draw"
	"image/jpeg"
	"os"

	xdraw "golang.org/x/image/draw"
	_ "golang.org/x/image/webp"
)

// mergeImages merges a background image with an overlay.
func mergeImages(bgData, ovData []byte, outPath string) {
	bgImg, _, _ := image.Decode(bytes.NewReader(bgData))
	ovImg, _, _ := image.Decode(bytes.NewReader(ovData))
	if bgImg == nil || ovImg == nil {
		return
	}
	bounds := bgImg.Bounds()
	final := image.NewRGBA(bounds)
	draw.Draw(final, bounds, bgImg, image.Point{}, draw.Src)
	resizedOv := image.NewRGBA(bounds)
	xdraw.BiLinear.Scale(resizedOv, bounds, ovImg, ovImg.Bounds(), xdraw.Over, nil)
	draw.Draw(final, bounds, resizedOv, image.Point{}, draw.Over)
	f, _ := os.Create(outPath)
	defer f.Close()
	jpeg.Encode(f, final, &jpeg.Options{Quality: 90})
}
