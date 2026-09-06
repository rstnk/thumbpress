# thumbpress

`thumbpress` creates a YouTube-ready 1280x720 thumbnail from a background image, title, and optional subtitle.

## Build

```text
go build -o ./thumbpress ./cmd/thumbpress
```

## Render a thumbnail

```text
./thumbpress render \
  --input background.jpg \
  --output thumbnail.jpg \
  --title "Build better thumbnails" \
  --subtitle "A Go CLI walkthrough" \
  --font anton
```

The command accepts JPEG, PNG, and WebP background images up to 50 megapixels. Output uses the extension you choose: `.jpg`, `.jpeg`, or `.png`.

The title is placed in the top-left. The optional subtitle is placed in the bottom-right. Both use white text with a dark outline and shadow. A dark edge gradient protects the text while leaving the centre of the background visible.

## Options

```text
--input PATH       Background image path. Required.
--output PATH      Output .jpg, .jpeg, or .png path. Required.
--title TEXT       Title text. Required.
--subtitle TEXT    Optional subtitle text.
--font NAME        anton, archivo-black, bebas-neue, or inter. Default: anton.
--quality NUMBER   JPEG quality from 1 through 100. Default: 90.
```

Titles automatically wrap and shrink to fit three lines. Subtitles automatically wrap and shrink to fit two lines. `thumbpress` returns a clear error when text still cannot fit. The subtitle uses embedded Inter for readability.

`--output` must differ from `--input`. An existing output file is replaced only after encoding succeeds.

## Fonts and licenses

The binary embeds Anton, Archivo Black, Bebas Neue, and Inter, so it does not depend on fonts installed on the machine. Every font is distributed under the SIL Open Font License, Version 1.1. See [THIRD_PARTY_NOTICES](THIRD_PARTY_NOTICES) and the license text stored next to each asset in `assets/fonts/`.

## Verify

```text
go test ./...
go vet ./...
```
