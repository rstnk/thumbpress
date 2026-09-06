package render

import (
	"image"
	"image/color"
	"image/draw"
	"testing"
)

func TestCropToCanvasUsesCenteredCoverCrop(t *testing.T) {
	source := image.NewRGBA(image.Rect(0, 0, 2400, 720))
	draw.Draw(source, source.Bounds(), image.NewUniform(color.RGBA{R: 255, A: 255}), image.Point{}, draw.Src)
	draw.Draw(source, image.Rect(560, 0, 1840, 720), image.NewUniform(color.RGBA{G: 255, A: 255}), image.Point{}, draw.Src)

	canvas := CropToCanvas(source)
	if canvas.Bounds() != canvasBounds {
		t.Fatalf("canvas bounds = %v, want %v", canvas.Bounds(), canvasBounds)
	}

	for _, point := range []image.Point{{0, 0}, {CanvasWidth / 2, CanvasHeight / 2}, {CanvasWidth - 1, CanvasHeight - 1}} {
		got := color.RGBAModel.Convert(canvas.At(point.X, point.Y)).(color.RGBA)
		if got.G < 250 || got.R > 5 {
			t.Errorf("pixel at %v = %#v, want centered green crop", point, got)
		}
	}
}

func TestApplyEdgeOverlayDarkensBothEdges(t *testing.T) {
	canvas := image.NewRGBA(canvasBounds)
	draw.Draw(canvas, canvasBounds, image.NewUniform(color.RGBA{R: 255, G: 255, B: 255, A: 255}), image.Point{}, draw.Src)

	ApplyEdgeOverlay(canvas)

	top := color.RGBAModel.Convert(canvas.At(CanvasWidth/2, 0)).(color.RGBA)
	center := color.RGBAModel.Convert(canvas.At(CanvasWidth/2, CanvasHeight/2)).(color.RGBA)
	bottom := color.RGBAModel.Convert(canvas.At(CanvasWidth/2, CanvasHeight-1)).(color.RGBA)
	if top.R >= center.R {
		t.Errorf("top red = %d, center red = %d; want a darker top edge", top.R, center.R)
	}
	if bottom.R >= center.R {
		t.Errorf("bottom red = %d, center red = %d; want a darker bottom edge", bottom.R, center.R)
	}
	if center.R != 255 {
		t.Errorf("center red = %d, want 255", center.R)
	}
}

func TestCropToCanvasCoversCommonAspectRatios(t *testing.T) {
	tests := []struct {
		name   string
		width  int
		height int
	}{
		{name: "landscape", width: 2400, height: 720},
		{name: "portrait", width: 720, height: 2400},
		{name: "square", width: 1000, height: 1000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := image.NewRGBA(image.Rect(0, 0, tt.width, tt.height))
			draw.Draw(source, source.Bounds(), image.NewUniform(color.RGBA{R: 120, G: 120, B: 120, A: 255}), image.Point{}, draw.Src)

			canvas := CropToCanvas(source)
			if canvas.Bounds() != canvasBounds {
				t.Fatalf("canvas bounds = %v, want %v", canvas.Bounds(), canvasBounds)
			}
			for _, point := range []image.Point{{0, 0}, {CanvasWidth - 1, CanvasHeight - 1}} {
				if alpha := color.RGBAModel.Convert(canvas.At(point.X, point.Y)).(color.RGBA).A; alpha != 255 {
					t.Errorf("pixel alpha at %v = %d, want 255", point, alpha)
				}
			}
		})
	}
}
