package render

import (
	"image"
	"image/color"
	"strings"
	"testing"

	"github.com/rstnk/thumbpress/internal/fonts"
)

func TestRenderThumbnailDrawsTitleAndSubtitle(t *testing.T) {
	titleFont, err := fonts.Open(fonts.Anton)
	if err != nil {
		t.Fatalf("opening title font: %v", err)
	}
	subtitleFont, err := fonts.Open(fonts.Inter)
	if err != nil {
		t.Fatalf("opening subtitle font: %v", err)
	}

	source := image.NewRGBA(image.Rect(0, 0, CanvasWidth, CanvasHeight))
	for y := range CanvasHeight {
		for x := range CanvasWidth {
			source.SetRGBA(x, y, color.RGBA{R: 80, G: 80, B: 80, A: 255})
		}
	}

	canvas, err := RenderThumbnail(source, TextConfig{
		Title:        "Build better thumbnails",
		Subtitle:     "A Go CLI walkthrough",
		TitleFont:    titleFont,
		SubtitleFont: subtitleFont,
	})
	if err != nil {
		t.Fatalf("RenderThumbnail() error = %v", err)
	}
	if !hasBrightPixel(canvas, titleArea) {
		t.Error("title area contains no white text pixels")
	}
	if !hasBrightPixel(canvas, subtitleArea) {
		t.Error("subtitle area contains no white text pixels")
	}
}

func TestLayoutTextRightAlignsLines(t *testing.T) {
	parsed, err := fonts.Open(fonts.Inter)
	if err != nil {
		t.Fatalf("opening font: %v", err)
	}

	layout, err := layoutText("A short subtitle", textOptions{
		area:      subtitleArea,
		alignment: alignRight,
		font:      parsed,
		maxLines:  2,
		maxSize:   subtitleMaxSize,
		minSize:   subtitleMinSize,
		name:      "subtitle",
	})
	if err != nil {
		t.Fatalf("layoutText() error = %v", err)
	}
	defer layout.face.Close()

	for _, line := range layout.lines {
		if line.x+line.width != subtitleArea.Max.X {
			t.Errorf("line right edge = %d, want %d", line.x+line.width, subtitleArea.Max.X)
		}
	}
}

func TestLayoutTextWrapsLongTitle(t *testing.T) {
	parsed, err := fonts.Open(fonts.Anton)
	if err != nil {
		t.Fatalf("opening font: %v", err)
	}

	layout, err := layoutText("Build better thumbnails", textOptions{
		area:      titleArea,
		alignment: alignLeft,
		font:      parsed,
		maxLines:  3,
		maxSize:   titleMaxSize,
		minSize:   titleMinSize,
		name:      "title",
	})
	if err != nil {
		t.Fatalf("layoutText() error = %v", err)
	}
	defer layout.face.Close()

	if len(layout.lines) < 2 {
		t.Errorf("line count = %d, want a wrapped title", len(layout.lines))
	}
}

func TestRenderThumbnailRejectsUnfitTitle(t *testing.T) {
	titleFont, err := fonts.Open(fonts.Anton)
	if err != nil {
		t.Fatalf("opening title font: %v", err)
	}

	_, err = RenderThumbnail(image.NewRGBA(canvasBounds), TextConfig{
		Title:     strings.Repeat("W", 200),
		TitleFont: titleFont,
	})
	if err == nil || !strings.Contains(err.Error(), "title text does not fit") {
		t.Fatalf("RenderThumbnail() error = %v, want title fit error", err)
	}
}

func hasBrightPixel(canvas *image.RGBA, area image.Rectangle) bool {
	for y := area.Min.Y; y < area.Max.Y; y++ {
		for x := area.Min.X; x < area.Max.X; x++ {
			pixel := color.RGBAModel.Convert(canvas.At(x, y)).(color.RGBA)
			if pixel.R > 240 && pixel.G > 240 && pixel.B > 240 {
				return true
			}
		}
	}
	return false
}
