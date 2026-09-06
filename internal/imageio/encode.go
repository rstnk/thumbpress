package imageio

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"
)

// EncodeFile writes image to path as JPEG or PNG based on its extension.
func EncodeFile(path string, source image.Image, quality int) error {
	if source == nil {
		return fmt.Errorf("output image is required")
	}

	extension := strings.ToLower(filepath.Ext(path))
	if extension != ".jpg" && extension != ".jpeg" && extension != ".png" {
		return fmt.Errorf("unsupported output extension %q; use .jpg, .jpeg, or .png", extension)
	}
	if (extension == ".jpg" || extension == ".jpeg") && (quality < 1 || quality > 100) {
		return fmt.Errorf("JPEG quality must be between 1 and 100")
	}

	directory := filepath.Dir(path)
	temporary, err := os.CreateTemp(directory, ".thumbpress-*")
	if err != nil {
		return fmt.Errorf("creating output file for %q: %w", path, err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)

	if extension == ".png" {
		err = png.Encode(temporary, source)
	} else {
		err = jpeg.Encode(temporary, source, &jpeg.Options{Quality: quality})
	}
	if err != nil {
		temporary.Close()
		return fmt.Errorf("encoding output image %q: %w", path, err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("closing output image %q: %w", path, err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("finalizing output image %q: %w", path, err)
	}

	return nil
}
