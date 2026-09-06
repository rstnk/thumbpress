package cli

import (
	"bytes"
	"strings"
	"testing"

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
