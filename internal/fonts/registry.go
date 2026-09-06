// Package fonts defines the supported embedded font choices.
package fonts

import (
	"fmt"

	assets "github.com/rstnk/thumbpress/assets/fonts"
	"golang.org/x/image/font/opentype"
)

// Name identifies a bundled display font.
type Name string

const (
	Anton        Name = "anton"
	ArchivoBlack Name = "archivo-black"
	BebasNeue    Name = "bebas-neue"
	Inter        Name = "inter"
)

const Default = Anton

var fontFiles = map[Name]string{
	Anton:        "anton/Anton-Regular.ttf",
	ArchivoBlack: "archivo-black/ArchivoBlack-Regular.ttf",
	BebasNeue:    "bebas-neue/BebasNeue-Regular.ttf",
	Inter:        "inter/Inter-Variable.ttf",
}

// Parse validates a user-provided embedded font name.
func Parse(value string) (Name, error) {
	name := Name(value)
	if _, ok := fontFiles[name]; !ok {
		return "", fmt.Errorf("unsupported font %q; choose anton, archivo-black, bebas-neue, or inter", value)
	}

	return name, nil
}

// File returns the embedded font bytes for name.
func File(name Name) ([]byte, error) {
	path, ok := fontFiles[name]
	if !ok {
		return nil, fmt.Errorf("unsupported font %q", name)
	}

	data, err := assets.Files.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading embedded font %q: %w", name, err)
	}

	return data, nil
}

// Open parses an embedded font for use by the renderer.
func Open(name Name) (*opentype.Font, error) {
	data, err := File(name)
	if err != nil {
		return nil, err
	}

	parsed, err := opentype.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("parsing embedded font %q: %w", name, err)
	}

	return parsed, nil
}

// Names returns the supported font names in command-line order.
func Names() []Name {
	return []Name{Anton, ArchivoBlack, BebasNeue, Inter}
}
