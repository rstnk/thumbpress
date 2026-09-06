// Package render composes thumbnail imagery onto a fixed canvas.
package render

import (
	"image"
	"image/color"
	stdDraw "image/draw"
	"math"

	xdraw "golang.org/x/image/draw"
)

const (
	CanvasWidth  = 1280
	CanvasHeight = 720

	edgeGradientHeight = 220
	edgeGradientAlpha  = 150
)

var canvasBounds = image.Rect(0, 0, CanvasWidth, CanvasHeight)

// PrepareBackground center-crops source, scales it to the thumbnail canvas,
// and applies the default dark edge overlay.
func PrepareBackground(source image.Image) *image.RGBA {
	canvas := CropToCanvas(source)
	ApplyEdgeOverlay(canvas)
	return canvas
}

// CropToCanvas center-crops source to the canvas aspect ratio and scales it.
func CropToCanvas(source image.Image) *image.RGBA {
	canvas := image.NewRGBA(canvasBounds)
	stdDraw.Draw(canvas, canvasBounds, image.NewUniform(color.Black), image.Point{}, stdDraw.Src)

	xdraw.CatmullRom.Scale(canvas, canvasBounds, source, coverRect(source.Bounds()), xdraw.Over, nil)
	return canvas
}

// ApplyEdgeOverlay darkens the top and bottom edges of canvas in place.
func ApplyEdgeOverlay(canvas *image.RGBA) {
	bounds := canvas.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		alpha := edgeAlpha(y - bounds.Min.Y)
		if alpha == 0 {
			continue
		}

		multiplier := uint16(255 - alpha)
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			offset := canvas.PixOffset(x, y)
			canvas.Pix[offset] = uint8(uint16(canvas.Pix[offset]) * multiplier / 255)
			canvas.Pix[offset+1] = uint8(uint16(canvas.Pix[offset+1]) * multiplier / 255)
			canvas.Pix[offset+2] = uint8(uint16(canvas.Pix[offset+2]) * multiplier / 255)
		}
	}
}

func coverRect(bounds image.Rectangle) image.Rectangle {
	width := bounds.Dx()
	height := bounds.Dy()
	if width == 0 || height == 0 {
		return bounds
	}

	sourceRatio := float64(width) / float64(height)
	canvasRatio := float64(CanvasWidth) / float64(CanvasHeight)
	if sourceRatio > canvasRatio {
		cropWidth := int(math.Round(float64(height) * canvasRatio))
		x := bounds.Min.X + (width-cropWidth)/2
		return image.Rect(x, bounds.Min.Y, x+cropWidth, bounds.Max.Y)
	}

	cropHeight := int(math.Round(float64(width) / canvasRatio))
	y := bounds.Min.Y + (height-cropHeight)/2
	return image.Rect(bounds.Min.X, y, bounds.Max.X, y+cropHeight)
}

func edgeAlpha(y int) uint8 {
	topStrength := edgeStrength(y)
	bottomStrength := edgeStrength(CanvasHeight - 1 - y)
	strength := math.Max(topStrength, bottomStrength)
	return uint8(math.Round(strength * strength * edgeGradientAlpha))
}

func edgeStrength(distance int) float64 {
	if distance >= edgeGradientHeight {
		return 0
	}
	return float64(edgeGradientHeight-distance) / edgeGradientHeight
}
