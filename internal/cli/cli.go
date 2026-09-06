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
	"github.com/rstnk/thumbpress/internal/manifest"
	"github.com/rstnk/thumbpress/internal/render"
)

const usage = `Usage:
  thumbpress render [flags]

Commands:
  render    Create a thumbnail from a background image and title text.
  batch     Create thumbnails described by a JSON or CSV manifest.
`

// RenderOptions contains the inputs needed to render one thumbnail.
type RenderOptions struct {
	Input    string
	Output   string
	Title    string
	Subtitle string
	Font     fonts.Name
	Quality  int
}

// BatchOptions contains the command-line inputs for a batch render.
type BatchOptions struct {
	Manifest string
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
	case "batch":
		return runBatch(args[1:], stdout, stderr)
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

	if err := renderThumbnail(options); err != nil {
		fmt.Fprintf(stderr, "render: %v\n", err)
		return 1
	}
	return 0
}

func runBatch(args []string, stdout, stderr io.Writer) int {
	options, err := ParseBatchOptions(args, stdout, stderr)
	if errors.Is(err, flag.ErrHelp) {
		return 0
	}
	if err != nil {
		fmt.Fprintf(stderr, "batch: %v\n", err)
		return 2
	}

	jobs, err := manifest.Load(options.Manifest)
	if err != nil {
		fmt.Fprintf(stderr, "batch: %v\n", err)
		return 1
	}
	renderOptions, err := batchRenderOptions(jobs)
	if err != nil {
		fmt.Fprintf(stderr, "batch: %v\n", err)
		return 1
	}

	failures := 0
	for index, item := range renderOptions {
		if err := renderThumbnail(item); err != nil {
			fmt.Fprintf(stderr, "batch: job %d: %v\n", jobs[index].Row, err)
			failures++
			continue
		}
		fmt.Fprintf(stdout, "rendered %s\n", item.Output)
	}
	if failures > 0 {
		fmt.Fprintf(stderr, "batch: %d of %d jobs failed\n", failures, len(renderOptions))
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

	font, err := fonts.Parse(*fontName)
	if err != nil {
		return RenderOptions{}, err
	}
	options.Font = font
	if err := validateRenderOptions(options); err != nil {
		return RenderOptions{}, err
	}

	return options, nil
}

// ParseBatchOptions parses and validates batch command flags.
func ParseBatchOptions(args []string, stdout, stderr io.Writer) (BatchOptions, error) {
	var options BatchOptions
	flags := flag.NewFlagSet("batch", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.StringVar(&options.Manifest, "manifest", "", "JSON or CSV manifest path (required)")
	flags.Usage = func() {
		fmt.Fprint(stdout, `Usage:
  thumbpress batch --manifest FILE

The manifest must be a .json or .csv file.

Flags:
`)
		flags.PrintDefaults()
	}
	if err := flags.Parse(args); err != nil {
		return BatchOptions{}, err
	}
	if flags.NArg() > 0 {
		return BatchOptions{}, fmt.Errorf("unexpected argument %q", flags.Arg(0))
	}
	if strings.TrimSpace(options.Manifest) == "" {
		return BatchOptions{}, errors.New("--manifest is required")
	}
	return options, nil
}

func batchRenderOptions(jobs []manifest.Job) ([]RenderOptions, error) {
	options := make([]RenderOptions, 0, len(jobs))
	outputs := make(map[string]int, len(jobs))
	for _, job := range jobs {
		item, err := renderOptionsForJob(job)
		if err != nil {
			return nil, fmt.Errorf("job %d: %w", job.Row, err)
		}
		outputPath, err := filepath.Abs(item.Output)
		if err != nil {
			return nil, fmt.Errorf("job %d: resolving output path %q: %w", job.Row, item.Output, err)
		}
		if previous, exists := outputs[outputPath]; exists {
			return nil, fmt.Errorf("job %d: output path %q is already used by job %d", job.Row, item.Output, previous)
		}
		outputs[outputPath] = job.Row
		options = append(options, item)
	}
	return options, nil
}

func renderOptionsForJob(job manifest.Job) (RenderOptions, error) {
	fontName := string(fonts.Default)
	if strings.TrimSpace(job.Font) != "" {
		fontName = job.Font
	}
	font, err := fonts.Parse(fontName)
	if err != nil {
		return RenderOptions{}, err
	}
	quality := 90
	if job.Quality != nil {
		quality = *job.Quality
	}
	options := RenderOptions{
		Input:    job.Input,
		Output:   job.Output,
		Title:    job.Title,
		Subtitle: job.Subtitle,
		Font:     font,
		Quality:  quality,
	}
	if err := validateRenderOptions(options); err != nil {
		return RenderOptions{}, err
	}
	if err := validateDistinctPaths(options.Input, options.Output); err != nil {
		return RenderOptions{}, err
	}
	return options, nil
}

func validateRenderOptions(options RenderOptions) error {
	if strings.TrimSpace(options.Input) == "" {
		return errors.New("--input is required")
	}
	if strings.TrimSpace(options.Output) == "" {
		return errors.New("--output is required")
	}
	if strings.TrimSpace(options.Title) == "" {
		return errors.New("--title is required")
	}
	if options.Quality < 1 || options.Quality > 100 {
		return errors.New("--quality must be between 1 and 100")
	}
	return nil
}

func renderThumbnail(options RenderOptions) error {
	background, err := imageio.DecodeFile(options.Input)
	if err != nil {
		return err
	}
	titleFont, err := fonts.Open(options.Font)
	if err != nil {
		return err
	}

	config := render.TextConfig{Title: options.Title, TitleFont: titleFont}
	if strings.TrimSpace(options.Subtitle) != "" {
		subtitleFont, err := fonts.Open(fonts.Inter)
		if err != nil {
			return err
		}
		config.Subtitle = options.Subtitle
		config.SubtitleFont = subtitleFont
	}

	thumbnail, err := render.RenderThumbnail(background, config)
	if err != nil {
		return err
	}
	if err := imageio.EncodeFile(options.Output, thumbnail, options.Quality); err != nil {
		return err
	}
	return nil
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
