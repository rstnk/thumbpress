package render

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"math"
	"strings"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

const (
	titleMaxSize    = 116
	titleMinSize    = 40
	subtitleMaxSize = 48
	subtitleMinSize = 24

	textStrokeRadius = 3
	textShadowOffset = 4
)

var (
	titleArea    = image.Rect(36, 16, 804, 328)
	subtitleArea = image.Rect(488, 504, 1208, 656)
)

// TextConfig holds the type treatments for a thumbnail.
type TextConfig struct {
	Title        string
	Subtitle     string
	TitleFont    *opentype.Font
	SubtitleFont *opentype.Font
}

type textAlignment int

const (
	alignLeft textAlignment = iota
	alignRight
)

type textOptions struct {
	area      image.Rectangle
	alignment textAlignment
	font      *opentype.Font
	maxLines  int
	maxSize   int
	minSize   int
	name      string
}

type textLine struct {
	text  string
	x     int
	y     int
	width int
}

type textLayout struct {
	face  font.Face
	lines []textLine
}

// RenderThumbnail prepares source and draws the configured text blocks.
func RenderThumbnail(source image.Image, config TextConfig) (*image.RGBA, error) {
	if source == nil {
		return nil, errors.New("background image is required")
	}
	if strings.TrimSpace(config.Title) == "" {
		return nil, errors.New("title text is required")
	}
	if config.TitleFont == nil {
		return nil, errors.New("title font is required")
	}
	if strings.TrimSpace(config.Subtitle) != "" && config.SubtitleFont == nil {
		return nil, errors.New("subtitle font is required when subtitle text is set")
	}

	canvas := PrepareBackground(source)
	title, err := layoutText(config.Title, textOptions{
		area:      titleArea,
		alignment: alignLeft,
		font:      config.TitleFont,
		maxLines:  3,
		maxSize:   titleMaxSize,
		minSize:   titleMinSize,
		name:      "title",
	})
	if err != nil {
		return nil, err
	}
	defer title.face.Close()
	drawLayout(canvas, title)

	if strings.TrimSpace(config.Subtitle) == "" {
		return canvas, nil
	}

	subtitle, err := layoutText(config.Subtitle, textOptions{
		area:      subtitleArea,
		alignment: alignRight,
		font:      config.SubtitleFont,
		maxLines:  2,
		maxSize:   subtitleMaxSize,
		minSize:   subtitleMinSize,
		name:      "subtitle",
	})
	if err != nil {
		return nil, err
	}
	defer subtitle.face.Close()
	drawLayout(canvas, subtitle)

	return canvas, nil
}

func layoutText(value string, options textOptions) (textLayout, error) {
	for size := options.maxSize; size >= options.minSize; size -= 2 {
		face, err := opentype.NewFace(options.font, &opentype.FaceOptions{
			Size:    float64(size),
			DPI:     72,
			Hinting: font.HintingFull,
		})
		if err != nil {
			return textLayout{}, fmt.Errorf("creating %s font face: %w", options.name, err)
		}

		lines := wrapText(value, face, options.area.Dx())
		metrics := face.Metrics()
		ascent := metrics.Ascent.Ceil()
		descent := metrics.Descent.Ceil()
		lineHeight := int(math.Ceil(float64(ascent+descent) * 0.75))
		height := ascent + descent + (len(lines)-1)*lineHeight

		if len(lines) <= options.maxLines && height <= options.area.Dy() && linesFit(lines, face, options.area.Dx()) {
			return textLayout{face: face, lines: positionLines(lines, face, options.area, options.alignment, lineHeight, ascent, descent)}, nil
		}
		face.Close()
	}

	return textLayout{}, fmt.Errorf("%s text does not fit; shorten it to at most %d lines", options.name, options.maxLines)
}

func wrapText(value string, face font.Face, maxWidth int) []string {
	var lines []string
	for paragraph := range strings.SplitSeq(value, "\n") {
		words := strings.Fields(paragraph)
		if len(words) == 0 {
			continue
		}

		line := words[0]
		for _, word := range words[1:] {
			candidate := line + " " + word
			if measure(candidate, face) <= maxWidth {
				line = candidate
				continue
			}
			lines = append(lines, line)
			line = word
		}
		lines = append(lines, line)
	}
	return lines
}

func linesFit(lines []string, face font.Face, maxWidth int) bool {
	for _, line := range lines {
		if measure(line, face) > maxWidth {
			return false
		}
	}
	return true
}

func positionLines(lines []string, face font.Face, area image.Rectangle, alignment textAlignment, lineHeight, ascent, descent int) []textLine {
	positioned := make([]textLine, 0, len(lines))
	baseline := area.Min.Y + ascent
	if alignment == alignRight {
		baseline = area.Max.Y - descent - (len(lines)-1)*lineHeight
	}

	for index, line := range lines {
		width := measure(line, face)
		x := area.Min.X
		if alignment == alignRight {
			x = area.Max.X - width
		}
		positioned = append(positioned, textLine{
			text:  line,
			x:     x,
			y:     baseline + index*lineHeight,
			width: width,
		})
	}
	return positioned
}

func measure(value string, face font.Face) int {
	drawer := font.Drawer{Face: face}
	return drawer.MeasureString(value).Ceil()
}

func drawLayout(canvas *image.RGBA, layout textLayout) {
	for _, line := range layout.lines {
		drawString(canvas, layout.face, line.x+textShadowOffset, line.y+textShadowOffset, line.text, color.RGBA{A: 190})
		for y := -textStrokeRadius; y <= textStrokeRadius; y++ {
			for x := -textStrokeRadius; x <= textStrokeRadius; x++ {
				if x*x+y*y > textStrokeRadius*textStrokeRadius {
					continue
				}
				drawString(canvas, layout.face, line.x+x, line.y+y, line.text, color.RGBA{A: 220})
			}
		}
		drawString(canvas, layout.face, line.x, line.y, line.text, color.RGBA{R: 255, G: 255, B: 255, A: 255})
	}
}

func drawString(canvas *image.RGBA, face font.Face, x, y int, value string, source color.RGBA) {
	drawer := font.Drawer{
		Dst:  canvas,
		Src:  image.NewUniform(source),
		Face: face,
		Dot:  fixed.P(x, y),
	}
	drawer.DrawString(value)
}
