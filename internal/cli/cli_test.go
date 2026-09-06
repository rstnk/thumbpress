package cli

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/rstnk/thumbpress/internal/fonts"
)

func TestParseRenderOptions(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	options, err := ParseRenderOptions([]string{
		"--input", "background.jpg",
		"--output", "thumbnail.jpg",
		"--title", "A title",
	}, stdout, stderr)
	if err != nil {
		t.Fatalf("ParseRenderOptions() error = %v", err)
	}

	if options.Font != fonts.Default {
		t.Errorf("Font = %q, want %q", options.Font, fonts.Default)
	}
	if options.Quality != 90 {
		t.Errorf("Quality = %d, want 90", options.Quality)
	}
}

func TestParseRenderOptionsGeneratesOutputPath(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	now := time.Date(2026, time.September, 6, 9, 12, 34, 0, time.UTC)

	options, err := parseRenderOptions([]string{
		"--input", "temp/o887mnz4jpk91.jpg",
		"--title", "A title",
	}, stdout, stderr, now)
	if err != nil {
		t.Fatalf("parseRenderOptions() error = %v", err)
	}

	if got, want := options.Output, "temp/o887mnz4jpk91_20260906091234.jpg"; got != want {
		t.Errorf("Output = %q, want %q", got, want)
	}
}

func TestParseRenderOptionsRejectsInvalidFont(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	_, err := ParseRenderOptions([]string{
		"--input", "background.jpg",
		"--output", "thumbnail.jpg",
		"--title", "A title",
		"--font", "invalid",
	}, stdout, stderr)
	if err == nil || !strings.Contains(err.Error(), "unsupported font") {
		t.Fatalf("ParseRenderOptions() error = %v, want unsupported font error", err)
	}
}

func TestRunRenderHelp(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	exitCode := Run([]string{"render", "--help"}, stdout, stderr)
	if exitCode != 0 {
		t.Fatalf("Run() exit code = %d, want 0", exitCode)
	}
	if !strings.Contains(stdout.String(), "--input") {
		t.Errorf("help output = %q, want input flag", stdout.String())
	}
}

func TestRunRenderWritesThumbnail(t *testing.T) {
	directory := t.TempDir()
	input := filepath.Join(directory, "background.jpg")
	output := filepath.Join(directory, "thumbnail.jpg")
	writeJPEGFixture(t, input)

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	exitCode := Run([]string{
		"render",
		"--input", input,
		"--output", output,
		"--title", "A title",
		"--subtitle", "A subtitle",
	}, stdout, stderr)
	if exitCode != 0 {
		t.Fatalf("Run() exit code = %d, stderr = %s", exitCode, stderr.String())
	}

	file, err := os.Open(output)
	if err != nil {
		t.Fatalf("opening output: %v", err)
	}
	defer file.Close()
	decoded, _, err := image.Decode(file)
	if err != nil {
		t.Fatalf("decoding output: %v", err)
	}
	if got, want := decoded.Bounds().Dx(), 1280; got != want {
		t.Errorf("output width = %d, want %d", got, want)
	}
	if got, want := decoded.Bounds().Dy(), 720; got != want {
		t.Errorf("output height = %d, want %d", got, want)
	}
}

func TestValidateDistinctPaths(t *testing.T) {
	path := filepath.Join(t.TempDir(), "background.jpg")
	if err := validateDistinctPaths(path, path); err == nil {
		t.Error("validateDistinctPaths() error = nil, want same-path error")
	}
}

func writeJPEGFixture(t *testing.T, path string) {
	t.Helper()
	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("creating fixture: %v", err)
	}
	defer file.Close()

	fixture := image.NewRGBA(image.Rect(0, 0, 1920, 1080))
	for y := range 1080 {
		for x := range 1920 {
			fixture.SetRGBA(x, y, color.RGBA{R: 90, G: 140, B: 180, A: 255})
		}
	}
	if err := jpeg.Encode(file, fixture, nil); err != nil {
		t.Fatalf("encoding fixture: %v", err)
	}
}
