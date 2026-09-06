package cli

import (
	"bytes"
	"encoding/json"
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

func TestRunBatchWritesThumbnails(t *testing.T) {
	directory := t.TempDir()
	writeJPEGFixture(t, filepath.Join(directory, "one.jpg"))
	writeJPEGFixture(t, filepath.Join(directory, "two.jpg"))
	if err := os.Mkdir(filepath.Join(directory, "output"), 0o700); err != nil {
		t.Fatalf("creating output directory: %v", err)
	}
	manifestPath := filepath.Join(directory, "jobs.json")
	writeJSONManifest(t, manifestPath, map[string]any{
		"jobs": []map[string]any{
			{"input": "one.jpg", "output": "output/one.jpg", "title": "First title"},
			{"input": "two.jpg", "output": "output/two.png", "title": "Second title", "subtitle": "Details", "font": "anton", "quality": 85},
		},
	})

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	exitCode := Run([]string{"batch", "--manifest", manifestPath}, stdout, stderr)
	if exitCode != 0 {
		t.Fatalf("Run() exit code = %d, stderr = %s", exitCode, stderr.String())
	}
	for _, filename := range []string{"one.jpg", "two.png"} {
		file, err := os.Open(filepath.Join(directory, "output", filename))
		if err != nil {
			t.Fatalf("opening %s: %v", filename, err)
		}
		decoded, _, err := image.Decode(file)
		file.Close()
		if err != nil {
			t.Fatalf("decoding %s: %v", filename, err)
		}
		if got, want := decoded.Bounds(), image.Rect(0, 0, 1280, 720); got != want {
			t.Errorf("%s bounds = %v, want %v", filename, got, want)
		}
	}
	if got, want := strings.Count(stdout.String(), "rendered"), 2; got != want {
		t.Errorf("rendered count = %d, want %d", got, want)
	}
}

func TestRunBatchWritesCSVThumbnail(t *testing.T) {
	directory := t.TempDir()
	writeJPEGFixture(t, filepath.Join(directory, "background.jpg"))
	if err := os.Mkdir(filepath.Join(directory, "output"), 0o700); err != nil {
		t.Fatalf("creating output directory: %v", err)
	}
	manifestPath := filepath.Join(directory, "jobs.csv")
	if err := os.WriteFile(manifestPath, []byte("input,output,title\nbackground.jpg,output/thumbnail.jpg,CSV title\n"), 0o600); err != nil {
		t.Fatalf("writing manifest: %v", err)
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	exitCode := Run([]string{"batch", "--manifest", manifestPath}, stdout, stderr)
	if exitCode != 0 {
		t.Fatalf("Run() exit code = %d, stderr = %s", exitCode, stderr.String())
	}
	if _, err := os.Stat(filepath.Join(directory, "output", "thumbnail.jpg")); err != nil {
		t.Fatalf("opening output: %v", err)
	}
	file, err := os.Open(filepath.Join(directory, "output", "thumbnail.jpg"))
	if err != nil {
		t.Fatalf("opening output for decoding: %v", err)
	}
	defer file.Close()
	decoded, _, err := image.Decode(file)
	if err != nil {
		t.Fatalf("decoding output: %v", err)
	}
	if got, want := decoded.Bounds(), image.Rect(0, 0, 1280, 720); got != want {
		t.Errorf("output bounds = %v, want %v", got, want)
	}
}

func TestRunBatchValidatesAllJobsBeforeRendering(t *testing.T) {
	directory := t.TempDir()
	writeJPEGFixture(t, filepath.Join(directory, "background.jpg"))
	manifestPath := filepath.Join(directory, "jobs.json")
	writeJSONManifest(t, manifestPath, map[string]any{
		"jobs": []map[string]any{
			{"input": "background.jpg", "output": "output/first.jpg", "title": "First title"},
			{"input": "background.jpg", "output": "output/second.jpg", "title": ""},
		},
	})

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	exitCode := Run([]string{"batch", "--manifest", manifestPath}, stdout, stderr)
	if exitCode != 1 {
		t.Fatalf("Run() exit code = %d, want 1; stderr = %s", exitCode, stderr.String())
	}
	if !strings.Contains(stderr.String(), "job 2") {
		t.Errorf("stderr = %q, want job number", stderr.String())
	}
	if _, err := os.Stat(filepath.Join(directory, "output", "first.jpg")); !os.IsNotExist(err) {
		t.Errorf("first output exists before validation finished, stat error = %v", err)
	}
}

func TestRunBatchContinuesAfterRenderingFailure(t *testing.T) {
	directory := t.TempDir()
	writeJPEGFixture(t, filepath.Join(directory, "background.jpg"))
	if err := os.Mkdir(filepath.Join(directory, "output"), 0o700); err != nil {
		t.Fatalf("creating output directory: %v", err)
	}
	manifestPath := filepath.Join(directory, "jobs.json")
	writeJSONManifest(t, manifestPath, map[string]any{
		"jobs": []map[string]any{
			{"input": "missing.jpg", "output": "output/missing.jpg", "title": "Missing background"},
			{"input": "background.jpg", "output": "output/finished.jpg", "title": "Finished thumbnail"},
		},
	})

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	exitCode := Run([]string{"batch", "--manifest", manifestPath}, stdout, stderr)
	if exitCode != 1 {
		t.Fatalf("Run() exit code = %d, want 1; stderr = %s", exitCode, stderr.String())
	}
	if !strings.Contains(stderr.String(), "job 1") {
		t.Errorf("stderr = %q, want failed job number", stderr.String())
	}
	if _, err := os.Stat(filepath.Join(directory, "output", "finished.jpg")); err != nil {
		t.Fatalf("later output was not written: %v", err)
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

func writeJSONManifest(t *testing.T, path string, value any) {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshaling manifest: %v", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("writing manifest: %v", err)
	}
}
