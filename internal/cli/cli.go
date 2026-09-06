// Package cli implements thumbpress command-line commands.
package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rstnk/thumbpress/internal/fonts"
	"github.com/rstnk/thumbpress/internal/imageio"
	"github.com/rstnk/thumbpress/internal/render"
)

const usage = `Usage:
  thumbpress render [flags]

Commands:
  render    Create a thumbnail from a background image and title text.
`

// RenderOptions contains the command-line inputs for a future render.
type RenderOptions struct {
	Input    string
	Output   string
	Title    string
	Subtitle string
	Font     fonts.Name
	Quality  int
}

// Run executes the thumbpress command and returns its process exit code.
func Run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 || isHelp(args[0]) {
		fmt.Fprint(stdout, usage)
		return 0
	}

	switch args[0] {
	case "render":
		return runRender(args[1:], stdout, stderr)
	default:
		fmt.Fprintf(stderr, "unknown command %q\n\n%s", args[0], usage)
		return 2
	}
}

func runRender(args []string, stdout, stderr io.Writer) int {
	options, err := ParseRenderOptions(args, stdout, stderr)
	if errors.Is(err, flag.ErrHelp) {
		return 0
	}
	if err != nil {
		fmt.Fprintf(stderr, "render: %v\n", err)
		return 2
	}

	if err := validateDistinctPaths(options.Input, options.Output); err != nil {
		fmt.Fprintf(stderr, "render: %v\n", err)
		return 2
	}

	background, err := imageio.DecodeFile(options.Input)
	if err != nil {
		fmt.Fprintf(stderr, "render: %v\n", err)
		return 1
	}
	titleFont, err := fonts.Open(options.Font)
	if err != nil {
		fmt.Fprintf(stderr, "render: %v\n", err)
		return 1
	}

	config := render.TextConfig{Title: options.Title, TitleFont: titleFont}
	if strings.TrimSpace(options.Subtitle) != "" {
		subtitleFont, err := fonts.Open(fonts.Inter)
		if err != nil {
			fmt.Fprintf(stderr, "render: %v\n", err)
			return 1
		}
		config.Subtitle = options.Subtitle
		config.SubtitleFont = subtitleFont
	}

	thumbnail, err := render.RenderThumbnail(background, config)
	if err != nil {
		fmt.Fprintf(stderr, "render: %v\n", err)
		return 1
	}
	if err := imageio.EncodeFile(options.Output, thumbnail, options.Quality); err != nil {
		fmt.Fprintf(stderr, "render: %v\n", err)
		return 1
	}

	return 0
}

// ParseRenderOptions parses and validates render command flags.
func ParseRenderOptions(args []string, stdout, stderr io.Writer) (RenderOptions, error) {
	return parseRenderOptions(args, stdout, stderr, time.Now())
}

func parseRenderOptions(args []string, stdout, stderr io.Writer, now time.Time) (RenderOptions, error) {
	options := RenderOptions{Font: fonts.Default, Quality: 90}

	flags := flag.NewFlagSet("render", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.StringVar(&options.Input, "input", "", "background image path (required)")
	flags.StringVar(&options.Output, "output", "", "output .jpg, .jpeg, or .png path")
	flags.StringVar(&options.Title, "title", "", "title text (required)")
	flags.StringVar(&options.Subtitle, "subtitle", "", "subtitle text")
	fontName := flags.String("font", string(fonts.Default), "font: anton, archivo-black, bebas-neue, or inter")
	flags.IntVar(&options.Quality, "quality", 90, "JPEG quality from 1 to 100")
	flags.Usage = func() {
		fmt.Fprint(stdout, `Usage:
  thumbpress render --input IMAGE --title TEXT [--output IMAGE] [flags]

Flags:
`)
		flags.PrintDefaults()
	}

	if err := flags.Parse(args); err != nil {
		return RenderOptions{}, err
	}
	if flags.NArg() > 0 {
		return RenderOptions{}, fmt.Errorf("unexpected argument %q", flags.Arg(0))
	}
	if strings.TrimSpace(options.Input) == "" {
		return RenderOptions{}, errors.New("--input is required")
	}
	if strings.TrimSpace(options.Title) == "" {
		return RenderOptions{}, errors.New("--title is required")
	}
	if strings.TrimSpace(options.Output) == "" {
		options.Output = defaultOutputPath(options.Input, now)
	}
	if options.Quality < 1 || options.Quality > 100 {
		return RenderOptions{}, errors.New("--quality must be between 1 and 100")
	}

	font, err := fonts.Parse(*fontName)
	if err != nil {
		return RenderOptions{}, err
	}
	options.Font = font

	return options, nil
}

func isHelp(arg string) bool {
	return arg == "-h" || arg == "--help" || arg == "help"
}

func validateDistinctPaths(input, output string) error {
	inputPath, err := filepath.Abs(input)
	if err != nil {
		return fmt.Errorf("resolving input path %q: %w", input, err)
	}
	outputPath, err := filepath.Abs(output)
	if err != nil {
		return fmt.Errorf("resolving output path %q: %w", output, err)
	}
	if inputPath == outputPath {
		return errors.New("--output must differ from --input")
	}

	inputInfo, err := os.Stat(inputPath)
	if err != nil {
		return nil
	}
	outputInfo, err := os.Stat(outputPath)
	if err == nil && os.SameFile(inputInfo, outputInfo) {
		return errors.New("--output must differ from --input")
	}

	return nil
}

func defaultOutputPath(input string, now time.Time) string {
	directory := filepath.Dir(input)
	base := filepath.Base(input)
	extension := filepath.Ext(base)
	name := strings.TrimSuffix(base, extension)
	if name == "" {
		name = base
	}
	return filepath.Join(directory, name+"_"+now.Format("20060102150405")+".jpg")
}
