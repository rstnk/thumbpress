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

const maxInputPixels = 50_000_000

// DecodeFile opens and decodes a supported background image.
func DecodeFile(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening input image %q: %w", path, err)
	}
	defer file.Close()

	config, _, err := image.DecodeConfig(file)
	if err != nil {
		return nil, fmt.Errorf("reading input image dimensions %q: %w", path, err)
	}
	if err := validateDimensions(config); err != nil {
		return nil, fmt.Errorf("validating input image %q: %w", path, err)
	}
	if _, err := file.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("rewinding input image %q: %w", path, err)
	}

	decoded, _, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("decoding input image %q: %w", path, err)
	}

	return decoded, nil
}

func validateDimensions(config image.Config) error {
	if config.Width <= 0 || config.Height <= 0 {
		return fmt.Errorf("image dimensions must be positive")
	}
	if int64(config.Width)*int64(config.Height) > maxInputPixels {
		return fmt.Errorf("image exceeds the %d-megapixel limit", maxInputPixels/1_000_000)
	}
	return nil
}
