// Package cli implements thumbpress command-line commands.
package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/rstnk/thumbpress/internal/fonts"
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

	_ = options
	fmt.Fprintln(stderr, "render: thumbnail rendering is not available yet")
	return 1
}

// ParseRenderOptions parses and validates render command flags.
func ParseRenderOptions(args []string, stdout, stderr io.Writer) (RenderOptions, error) {
	options := RenderOptions{Font: fonts.Default, Quality: 90}

	flags := flag.NewFlagSet("render", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.StringVar(&options.Input, "input", "", "background image path (required)")
	flags.StringVar(&options.Output, "output", "", "output .jpg, .jpeg, or .png path (required)")
	flags.StringVar(&options.Title, "title", "", "title text (required)")
	flags.StringVar(&options.Subtitle, "subtitle", "", "subtitle text")
	fontName := flags.String("font", string(fonts.Default), "font: anton, archivo-black, bebas-neue, or inter")
	flags.IntVar(&options.Quality, "quality", 90, "JPEG quality from 1 to 100")
	flags.Usage = func() {
		fmt.Fprint(stdout, `Usage:
  thumbpress render --input IMAGE --output IMAGE --title TEXT [flags]

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
	if strings.TrimSpace(options.Output) == "" {
		return RenderOptions{}, errors.New("--output is required")
	}
	if strings.TrimSpace(options.Title) == "" {
		return RenderOptions{}, errors.New("--title is required")
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
