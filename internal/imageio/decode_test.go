package imageio

import (
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestDecodeFile(t *testing.T) {
	source := image.NewRGBA(image.Rect(0, 0, 10, 5))
	source.SetRGBA(4, 2, color.RGBA{R: 255, A: 255})

	tests := []struct {
		name  string
		write func(string) error
	}{
		{
			name: "jpeg",
			write: func(path string) error {
				file, err := os.Create(path)
				if err != nil {
					return err
				}
				defer file.Close()
				return jpeg.Encode(file, source, nil)
			},
		},
		{
			name: "png",
			write: func(path string) error {
				file, err := os.Create(path)
				if err != nil {
					return err
				}
				defer file.Close()
				return png.Encode(file, source)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "input."+tt.name)
			if err := tt.write(path); err != nil {
				t.Fatalf("writing fixture: %v", err)
			}

			decoded, err := DecodeFile(path)
			if err != nil {
				t.Fatalf("DecodeFile() error = %v", err)
			}
			if decoded.Bounds() != source.Bounds() {
				t.Errorf("decoded bounds = %v, want %v", decoded.Bounds(), source.Bounds())
			}
		})
	}
}

func TestDecodeFileReturnsContextForInvalidImage(t *testing.T) {
	path := filepath.Join(t.TempDir(), "invalid.jpg")
	if err := os.WriteFile(path, []byte("not an image"), 0o600); err != nil {
		t.Fatalf("writing invalid fixture: %v", err)
	}

	if _, err := DecodeFile(path); err == nil {
		t.Fatal("DecodeFile() error = nil, want an error")
	}
}
