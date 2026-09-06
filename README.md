# thumbpress

`thumbpress` creates a YouTube-ready 1280x720 thumbnail from a background image, title, and optional subtitle.

## Build

```text
make build
```

## Render a thumbnail

```text
bin/thumbpress render \
  --input background.jpg \
  --output thumbnail.jpg \
  --title "Build better thumbnails" \
  --subtitle "A Go CLI walkthrough" \
  --font anton
```

The command accepts JPEG, PNG, and WebP background images up to 50 megapixels. Output uses the extension you choose: `.jpg`, `.jpeg`, or `.png`.

The title is placed in the top-left. The optional subtitle is placed in the bottom-right. Both use white text with a dark outline and shadow. A dark edge gradient protects the text while leaving the centre of the background visible.

## Options

| Flag | Required | Default | Behavior |
| --- | --- | --- | --- |
| `--input PATH` | Yes | None | Background image to decode. Supports JPEG, PNG, and WebP. |
| `--output PATH` | Yes | None | Output path ending in `.jpg`, `.jpeg`, or `.png`. It must differ from the input path. |
| `--title TEXT` | Yes | None | Title text placed in the top-left, automatically wrapped and sized to fit three lines. |
| `--subtitle TEXT` | No | Omitted | Adds a bottom-right subtitle, automatically wrapped and sized to fit two lines. |
| `--font NAME` | No | `anton` | Title font: `anton`, `archivo-black`, `bebas-neue`, or `inter`. |
| `--quality NUMBER` | No | `90` | JPEG quality from 1 through 100. PNG output ignores this flag. |

Titles automatically wrap and shrink to fit three lines. Subtitles automatically wrap and shrink to fit two lines. `thumbpress` returns a clear error when text still cannot fit. The subtitle uses embedded Inter for readability.

`--output` must differ from `--input`. An existing output file is replaced only after encoding succeeds.

## Fonts and licenses

The binary embeds Anton, Archivo Black, Bebas Neue, and Inter, so it does not depend on fonts installed on the machine. Every font is distributed under the SIL Open Font License, Version 1.1. See [THIRD_PARTY_NOTICES](THIRD_PARTY_NOTICES) and the license text stored next to each asset in `assets/fonts/`.

## Verify

```text
make test
make lint
```
