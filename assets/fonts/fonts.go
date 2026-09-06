// Package fonts exposes embedded font assets used by thumbpress.
package fonts

import "embed"

// Files contains the bundled fonts and their license texts.
//
//go:embed anton/Anton-Regular.ttf anton/OFL.txt archivo-black/ArchivoBlack-Regular.ttf archivo-black/OFL.txt bebas-neue/BebasNeue-Regular.ttf bebas-neue/OFL.txt inter/Inter-Variable.ttf inter/OFL.txt
var Files embed.FS
