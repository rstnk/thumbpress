// Package imageio decodes and encodes thumbnail image files.
package imageio

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"

	_ "golang.org/x/image/webp"
)

// DecodeFile opens and decodes a supported background image.
func DecodeFile(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening input image %q: %w", path, err)
	}
	defer file.Close()

	decoded, _, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("decoding input image %q: %w", path, err)
	}

	return decoded, nil
}
