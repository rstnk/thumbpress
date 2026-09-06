# thumbpress

`thumbpress` creates a YouTube-ready 1280x720 thumbnail from a background image, title, and optional subtitle.

## Build

```text
make build
```

## Install

```text
make install
```

This installs `thumbpress` to `~/.local/bin`. Set `INSTALL_DIR` to choose another location:

```text
make install INSTALL_DIR=/usr/local/bin
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

## Render a batch

Create a JSON or CSV manifest, then render every job in one command:

```text
bin/thumbpress batch --manifest thumbnails.json
```

JSON manifests contain a `jobs` array:

```json
{
  "jobs": [
    {
      "input": "backgrounds/episode-1.jpg",
      "output": "thumbnails/episode-1.jpg",
      "title": "Build better thumbnails",
      "subtitle": "A Go CLI walkthrough",
      "font": "anton",
      "quality": 90
    }
  ]
}
```

CSV manifests use this header:

```text
input,output,title,subtitle,font,quality
backgrounds/episode-1.jpg,thumbnails/episode-1.jpg,Build better thumbnails,A Go CLI walkthrough,anton,90
```

Every job requires `input`, `output`, and `title`. `subtitle` is optional. Omitted `font` and `quality` values use `bebas-neue` and `90`. Relative image paths resolve from the manifest's directory, and output directories must already exist. `thumbpress` validates every job before it renders any image. It continues after image-specific failures, reports each failure, and exits unsuccessfully when any job fails.

## Options

| Flag | Required | Default | Behavior |
| --- | --- | --- | --- |
| `--input PATH` | Yes | None | Background image to decode. Supports JPEG, PNG, and WebP. |
| `--output PATH` | No | `<input-base>_YYYYMMDDHHMMSS.jpg` | Output path ending in `.jpg`, `.jpeg`, or `.png`. When omitted, writes a JPEG beside the input using a local timestamp. |
| `--title TEXT` | Yes | None | Title text placed in the top-left, automatically wrapped and sized to fit three lines. |
| `--subtitle TEXT` | No | Omitted | Adds a bottom-right subtitle, automatically wrapped and sized to fit two lines. |
| `--font NAME` | No | `bebas-neue` | Title font: `anton`, `archivo-black`, `bebas-neue`, or `inter`. |
| `--quality NUMBER` | No | `90` | JPEG quality from 1 through 100. PNG output ignores this flag. |

`thumbpress batch` accepts one required option: `--manifest PATH`, ending in `.json` or `.csv`.

Titles automatically wrap and shrink to fit three lines. Subtitles automatically wrap and shrink to fit two lines. `thumbpress` returns a clear error when text still cannot fit. The subtitle uses embedded Inter for readability.

`--output` must differ from `--input`. An existing output file is replaced only after encoding succeeds.

## Fonts and licenses

The binary embeds Anton, Archivo Black, Bebas Neue, and Inter, so it does not depend on fonts installed on the machine. Every font is distributed under the SIL Open Font License, Version 1.1. See [THIRD_PARTY_NOTICES](THIRD_PARTY_NOTICES) and the license text stored next to each asset in `assets/fonts/`.

## Verify

```text
make test
make lint
```
