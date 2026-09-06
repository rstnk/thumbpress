package imageio

import (
	"image"
	"image/color"
	"path/filepath"
	"testing"
)

func TestEncodeFile(t *testing.T) {
	source := image.NewRGBA(image.Rect(0, 0, 8, 4))
	source.SetRGBA(2, 1, color.RGBA{R: 255, A: 255})

	for _, extension := range []string{".jpg", ".jpeg", ".png"} {
		t.Run(extension, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "thumbnail"+extension)
			if err := EncodeFile(path, source, 90); err != nil {
				t.Fatalf("EncodeFile() error = %v", err)
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

func TestEncodeFileRejectsInvalidOptions(t *testing.T) {
	source := image.NewRGBA(image.Rect(0, 0, 1, 1))
	directory := t.TempDir()

	if err := EncodeFile(filepath.Join(directory, "thumbnail.gif"), source, 90); err == nil {
		t.Error("EncodeFile() error = nil, want extension error")
	}
	if err := EncodeFile(filepath.Join(directory, "thumbnail.jpg"), source, 0); err == nil {
		t.Error("EncodeFile() error = nil, want quality error")
	}
}
