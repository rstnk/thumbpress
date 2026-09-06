package manifest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadJSON(t *testing.T) {
	path := writeManifest(t, "jobs.json", `{
  "jobs": [{
    "input": "images/background.jpg",
    "output": "out/thumbnail.png",
    "title": "Build better thumbnails",
    "subtitle": "A Go CLI walkthrough",
    "font": "anton",
    "quality": 85
  }]
}`)

	jobs, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got, want := jobs[0].Input, filepath.Join(filepath.Dir(path), "images/background.jpg"); got != want {
		t.Errorf("Input = %q, want %q", got, want)
	}
	if got, want := jobs[0].Output, filepath.Join(filepath.Dir(path), "out/thumbnail.png"); got != want {
		t.Errorf("Output = %q, want %q", got, want)
	}
	if jobs[0].Quality == nil || *jobs[0].Quality != 85 {
		t.Errorf("Quality = %v, want 85", jobs[0].Quality)
	}
	if jobs[0].Row != 1 {
		t.Errorf("Row = %d, want 1", jobs[0].Row)
	}
}

func TestLoadCSV(t *testing.T) {
	path := writeManifest(t, "jobs.csv", strings.Join([]string{
		"input,output,title,subtitle,font,quality",
		"background.jpg,out/one.jpg,First title,,anton,",
		"background.png,out/two.png,Second title,Details,inter,100",
	}, "\n"))

	jobs, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got, want := len(jobs), 2; got != want {
		t.Fatalf("job count = %d, want %d", got, want)
	}
	if jobs[0].Quality != nil {
		t.Errorf("first Quality = %v, want nil", jobs[0].Quality)
	}
	if jobs[1].Quality == nil || *jobs[1].Quality != 100 {
		t.Errorf("second Quality = %v, want 100", jobs[1].Quality)
	}
	if jobs[1].Row != 3 {
		t.Errorf("Row = %d, want 3", jobs[1].Row)
	}
}

func TestLoadCSVParsesQuotedValues(t *testing.T) {
	path := writeManifest(t, "jobs.csv", strings.Join([]string{
		"input,output,title,subtitle",
		"background.jpg,out.jpg,\"A title, with detail\",\"Line one",
		"line two\"",
	}, "\n"))

	jobs, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got, want := jobs[0].Title, "A title, with detail"; got != want {
		t.Errorf("Title = %q, want %q", got, want)
	}
	if got, want := jobs[0].Subtitle, "Line one\nline two"; got != want {
		t.Errorf("Subtitle = %q, want %q", got, want)
	}
}

func TestLoadRejectsInvalidManifest(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		contents string
		want     string
	}{
		{"unknown extension", "jobs.txt", "", "unsupported manifest extension"},
		{"unknown JSON field", "jobs.json", `{"jobs": [], "theme": "bold"}`, "unknown field"},
		{"missing CSV title", "jobs.csv", "input,output\nbackground.jpg,out.jpg", "missing required column"},
		{"invalid CSV quality", "jobs.csv", "input,output,title,quality\nbackground.jpg,out.jpg,A title,high", "invalid quality"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Load(writeManifest(t, tt.filename, tt.contents))
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("Load() error = %v, want %q", err, tt.want)
			}
		})
	}
}

func writeManifest(t *testing.T, filename, contents string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), filename)
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("writing manifest: %v", err)
	}
	return path
}
