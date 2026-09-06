# thumbpress specification

## Objective

`thumbpress` is a local Go command-line application that creates a YouTube-ready thumbnail from one background image, a title, and an optional subtitle. It produces a predictable 1280x720 image with readable white text and a simple, opinionated layout.

The initial audience is a creator who wants a repeatable way to produce clean thumbnails without opening a design application.

## Command contract

```text
thumbpress render \
  --input background.jpg \
  --output thumbnail.jpg \
  --title "Build better thumbnails" \
  --subtitle "A Go CLI walkthrough" \
  --font anton
```

Required flags:

- `--input`: path to a decodable JPEG, PNG, or WebP background image
- `--title`: non-empty title text

Optional flags:

- `--subtitle`: subtitle text
- `--output`: output `.jpg`, `.jpeg`, or `.png` path; defaults to `<input-base>_YYYYMMDDHHMMSS.jpg` beside the input
- `--font`: `anton`, `archivo-black`, `bebas-neue`, or `inter`; defaults to `bebas-neue`
- `--quality`: JPEG quality from 1 through 100; defaults to 90 and has no effect for PNG

`render` and `batch` are the supported subcommands. The first release does not expose layout, color, stroke, shadow, or overlay controls.

## Batch command contract

```text
thumbpress batch --manifest thumbnails.json
```

`--manifest` is required and accepts a `.json` or `.csv` file. A JSON manifest is an object with a `jobs` array. CSV manifests use a header row. Each job has these fields:

- `input`: required background image path
- `output`: required output image path
- `title`: required thumbnail title
- `subtitle`: optional thumbnail subtitle
- `font`: optional title font, defaulting to `bebas-neue`
- `quality`: optional JPEG quality, defaulting to `90`

Relative input and output paths resolve relative to the manifest. Batch validation checks every job's configuration, including duplicate output paths, before any render begins. Each render then follows the single-image rendering behavior. A failed job does not prevent remaining valid jobs from running; the command reports every failed job and exits unsuccessfully when one or more jobs fail.

## Rendering behavior

1. Decode the background image and use a centered cover crop to fill a 1280x720 canvas.
2. Apply a vertical dark overlay that is strongest at the top and bottom edges and fades toward the centre.
3. Draw the title in the top-left within a fixed safe area. It is white, left-aligned, wrapped, and sized down until it fits within three lines.
4. Draw an optional subtitle in the bottom-right within a fixed safe area. It is white, right-aligned, wrapped, and sized down until it fits within two lines.
5. Draw a dark outline and soft shadow beneath both text blocks.
6. Encode the result from the output extension. Do not overwrite the input file.

The title uses the selected face. The subtitle uses embedded Inter for legibility, independent of the title selection.

When text cannot fit at the minimum permitted size, the command exits with an error that asks the user to shorten that text.

## Font assets and licensing

The binary embeds a small set of static fonts with Go's `embed` package:

- Anton
- Archivo Black
- Bebas Neue
- Inter

Each bundled family must include its original Open Font License file in `assets/fonts/<family>/`. A generated or maintained `THIRD_PARTY_NOTICES` file must identify the fonts and their licenses. Font files are project assets, not system-font dependencies.

## Tech stack

- Go 1.26 or newer
- Go standard library for CLI parsing, image I/O, filesystem work, and JPEG/PNG encoding
- `golang.org/x/image` for font parsing, glyph drawing, and image transforms that are impractical in the standard library

The program must have no runtime network access and no platform-specific font lookup.

## Project structure

```text
assets/fonts/             Embedded fonts and their license files
cmd/thumbpress/           Thin executable entry point
internal/cli/             Commands, flags, and user-facing errors
internal/imageio/         Decode and output encoding
internal/render/          Cropping, overlay, layout, and text compositing
internal/fonts/           Embedded font registry
testdata/                 Small source fixtures and render references
docs/                     Product specification and implementation plan
THIRD_PARTY_NOTICES       Bundled-font notices
README.md                 Installation and usage
```

## Code style

Keep the command entry point small. Rendering code receives explicit configuration and returns a concrete image or a contextual error.

```go
func Render(background image.Image, config Config) (*image.RGBA, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("validating render configuration: %w", err)
	}

	canvas := cropToCanvas(background, canvasSize)
	return drawText(canvas, config)
}
```

Use `gofmt`, contextual error wrapping, table-driven unit tests, and interfaces only where they clarify a dependency. The command must avoid global mutable state.

## Commands

```text
make build
make install
make test
make lint
make fix
```

## Testing strategy

- Unit tests cover validation, cover-crop geometry, overlay interpolation, font selection, text bounds, and output-format selection.
- Integration tests execute the render pipeline with a small fixture image and an embedded font, then decode the result and verify its 1280x720 dimensions.
- Image assertions verify important pixels and bounds rather than brittle full-image byte equality.
- Manual verification uses `temp/o887mnz4jpk91.jpg`, which is 1920x1080, to inspect the full default layout.

## Boundaries

Always:

- Validate paths, extensions, required text, font identifiers, and output quality before rendering.
- Keep output deterministic for identical inputs and flags.
- Run formatting, tests, and vet checks before each implementation checkpoint.
- Preserve upstream font licenses with their assets and notices.

Ask first:

- Add dependencies beyond `golang.org/x/image`.
- Add templates, arbitrary positioning, logos, face detection, AI-generated imagery, or batch processing.
- Change the default output dimensions or fixed text layout.

Never:

- Depend on installed system fonts or runtime downloads.
- Send images or text across the network.
- Commit generated thumbnails, binaries, secrets, or unrelated temporary assets.

## Success criteria

- A user can render a JPEG or PNG thumbnail with the documented command and no external font installation.
- Output is exactly 1280x720 and decodes successfully as the extension's declared format.
- The title appears in the top-left and the optional subtitle in the bottom-right, with white fill, dark outline/shadow, and an edge gradient that protects both text blocks.
- Overlong text produces a useful error instead of clipped or off-canvas glyphs.
- The project builds, formats, passes tests, and passes `go vet ./...`.

## Deferred work

- User-defined themes and JSON configuration files
- Layout, color, and typography controls
- Multiple thumbnail templates
- SVG, GIF, and animated-image handling
