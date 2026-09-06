# thumbpress implementation plan

## Architecture

```text
CLI flags
    |
    v
validated render config ---> font registry
    |                          |
    v                          v
background decode ---> cover crop ---> edge overlay ---> text layout ---> encode
```

The CLI, image I/O, fonts, and renderer have one-way dependencies. This keeps layout and rendering unit-testable without invoking the executable.

## Task 1: Establish the executable and asset contract

Create the Go module, a minimal `thumbpress render` command, a package structure, `.gitignore`, README skeleton, embedded font registry, and license notices.

Acceptance criteria:

- [x] `go build ./cmd/thumbpress` produces an executable.
- [x] `thumbpress render --help` documents the approved flags.
- [x] Embedded font names are validated without system-font lookup.
- [x] Every bundled font has its OFL text recorded in the repository and notices file.

Verification:

- [x] `go build ./cmd/thumbpress`
- [x] `go test ./...`

Dependencies: None

Likely files:

- `go.mod`
- `cmd/thumbpress/main.go`
- `internal/cli/`
- `internal/fonts/`
- `assets/fonts/`
- `.gitignore`
- `THIRD_PARTY_NOTICES`
- `README.md`

Estimated scope: Medium

## Task 2: Build the background pipeline

Decode source images, center-crop them to the fixed canvas, and apply the dual-edge gradient overlay.

Acceptance criteria:

- [x] JPEG, PNG, and WebP backgrounds decode with contextual errors for unsupported or invalid input.
- [x] Every source aspect ratio produces a 1280x720 canvas without empty pixels.
- [x] The overlay darkens the top and bottom safe areas and leaves the centre visibly lighter.

Verification:

- [x] Unit tests cover centered cover crop and overlay alpha at top, centre, and bottom pixels.

Dependencies: Task 1

Likely files:

- `internal/imageio/`
- `internal/render/crop.go`
- `internal/render/overlay.go`
- `internal/render/*_test.go`

Estimated scope: Medium

## Task 3: Render title and subtitle

Implement safe-area text layout and compositing with automatic fit, white fill, dark outline, and shadow.

Acceptance criteria:

- [x] Titles render in the top-left and fit within three lines.
- [x] Subtitles render in the bottom-right and fit within two lines.
- [x] Both blocks stay inside their safe areas.
- [x] A clear validation error appears when either block cannot fit at the minimum font size.

Verification:

- [x] Tests cover alignment, minimum-size failure, and all font choices.
- [x] A manual render from the supplied fixture shows readable text against the background.

Dependencies: Tasks 1-2

Likely files:

- `internal/render/layout.go`
- `internal/render/text.go`
- `internal/render/style.go`
- `internal/render/*_test.go`

Estimated scope: Medium

## Checkpoint: End-to-end default render

- [x] The supplied fixture produces a readable 1280x720 thumbnail.
- [x] `go build ./cmd/thumbpress`, `go test ./...`, and `go vet ./...` pass.
- [x] Review the visual result before expanding output support.

## Task 4: Finish output and error handling

Add output encoding, JPEG quality validation, overwrite protection for the input path, and polished CLI errors.

Acceptance criteria:

- [x] `.jpg`, `.jpeg`, and `.png` output extensions choose the correct encoder.
- [x] JPEG quality accepts 1 through 100 and rejects values outside that range.
- [x] The command refuses an output path that resolves to the input image.
- [x] Failed writes and unsupported extensions identify the relevant path or flag.

Verification:

- [x] Integration tests decode generated JPEG and PNG outputs.
- [x] Tests cover invalid quality, invalid extension, and input-output collision errors.

Dependencies: Tasks 1-3

Likely files:

- `internal/imageio/encode.go`
- `internal/cli/render.go`
- `internal/imageio/*_test.go`
- `internal/cli/*_test.go`

Estimated scope: Small

## Task 5: Document and verify the release candidate

Complete installation and usage documentation, add a practical example, and run the full quality suite.

Acceptance criteria:

- [x] README covers installation, the rendering command, embedded-font choices, and output behavior.
- [x] The documentation contains no unsupported flags or features.
- [x] All test and static checks pass.

Verification:

- [x] `gofmt -w` reports no subsequent changes.
- [x] `go test ./...`
- [x] `go vet ./...`
- [x] Manually inspect a thumbnail made from `temp/o887mnz4jpk91.jpg`.

Dependencies: Tasks 1-4

Likely files:

- `README.md`
- `docs/SPEC.md`
- `docs/PLAN.md`

Estimated scope: Small

## Task 6: Render manifest batches

Add a `batch` subcommand that reads JSON and CSV manifests and renders each listed thumbnail through the existing pipeline.

Acceptance criteria:

- [x] JSON and CSV manifests accept `input`, `output`, `title`, `subtitle`, `font`, and `quality` fields.
- [x] Relative paths resolve from the manifest directory, and duplicate output paths are rejected before rendering.
- [x] All job configuration validates before the first output is written; later render failures are reported without skipping remaining jobs.
- [x] Documentation defines the manifest schema and batch failure behavior.

Verification:

- [x] Unit tests cover JSON and CSV parsing, including quoted CSV fields.
- [x] Integration tests render JSON and CSV manifest jobs to decodable 1280x720 files.
- [x] `go test ./...` and `go vet ./...`

Dependencies: Task 5

Likely files:

- `internal/manifest/`
- `internal/cli/`
- `README.md`
- `docs/SPEC.md`
- `docs/PLAN.md`

Estimated scope: Medium

## Risks and mitigations

| Risk | Impact | Mitigation |
| --- | --- | --- |
| Font acquisition or licensing ambiguity | High | Source each font from its official repository, retain its OFL text, and verify notices before embedding. |
| Text measurement differs from visual glyph bounds | Medium | Measure every wrapped line using the same font face and drawing parameters used for compositing. |
| Busy backgrounds reduce readability | Medium | Verify edge-overlay strength and text treatment against varied fixture images. |
| Unsupported image variants | Low | Return explicit decoding errors and document supported formats. |

## Implementation order

Tasks 1 through 5 are sequential. Task 2's crop and overlay tests can be prepared while the font assets are being collected, after Task 1 defines the shared package contracts.
