# Developer Tools Web Application Design

## Overview

A public utility website built with Go's Echo framework and HTMX, providing common developer debugging tools. Styled with Pico CSS for a clean, minimal appearance. Deployed via Docker using Chainguard images.

## Project Structure

```
tech-utils/
├── main.go                 # Entry point, config loading, server setup
├── handlers/
│   ├── home.go            # Landing page
│   ├── uuid.go            # UUID generation
│   ├── json.go            # JSON pretty-print + syntax highlight
│   ├── convert.go         # JSON/YAML/TOML conversions
│   ├── timezone.go        # Timezone converter
│   └── units.go           # Unit converter
├── templates/
│   ├── layout.html        # Base layout with nav, Pico CSS, HTMX
│   ├── home.html
│   ├── uuid.html
│   ├── json.html
│   ├── convert.html
│   ├── timezone.html
│   └── units.html
├── static/
│   └── syntax.css         # Minimal syntax highlighting colors
├── Dockerfile
└── go.mod
```

Templates are embedded in the binary using `//go:embed`. No external template files needed at runtime.

## Routes

```
GET  /                    → Home page with tool cards
GET  /uuid                → UUID generator page
POST /uuid/generate       → Returns generated UUIDs (HTMX partial)
GET  /json                → JSON pretty-printer page
POST /json/format         → Returns formatted JSON (HTMX partial)
GET  /convert             → Format converter page
POST /convert/transform   → Returns converted output (HTMX partial)
GET  /timezone            → Timezone converter page
POST /timezone/convert    → Returns converted times (HTMX partial)
GET  /units               → Unit converter page
POST /units/convert       → Returns converted value (HTMX partial)
```

## HTMX Pattern

Each tool page has a form. When submitted, HTMX posts to the action endpoint and swaps the result into a `<div id="result">` on the page. No full page reloads.

POST handlers detect HTMX requests via `HX-Request` header and return just the result fragment, not the full page.

Invalid input returns a styled error message in the result div rather than an HTTP error. The user stays on the page and can fix their input.

## Tool Specifications

### UUID Generator
- Input: Count (1-100), version (v4 random, v7 time-ordered)
- Output: List of UUIDs with copy button per line

### JSON Pretty-Printer
- Input: Raw JSON textarea, indent size (2/4 spaces or tabs)
- Output: Formatted JSON with syntax highlighting (keywords, strings, numbers, booleans colored via CSS classes)

### Format Converter
- Input: Source text, source format dropdown (JSON/YAML/TOML), target format dropdown
- Output: Converted text with syntax highlighting
- Validates input parses correctly before converting

### Timezone Converter
- Input: Datetime string, source timezone dropdown, target timezone dropdown
- Output: Converted datetime in multiple formats (ISO, RFC3339, Unix timestamp)
- Timezone list: Common zones (UTC, US timezones, Europe/London, Asia/Tokyo, etc.)

### Unit Converter
- Categories: Length, Weight, Temperature, Data size (bytes/KB/MB/GB)
- Input: Value, source unit, target unit
- Output: Converted value with formula shown

## Configuration

Environment variables with defaults:

| Variable | Default | Description |
|----------|---------|-------------|
| PORT | 8080 | Server port |
| LOG_LEVEL | info | debug, info, warn, error |
| ALLOWED_ORIGINS | * | CORS origins, comma-separated |

## Docker

Multi-stage build using Chainguard images:

```dockerfile
FROM cgr.dev/chainguard/go:latest AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o server .

FROM cgr.dev/chainguard/static:latest
COPY --from=builder /app/server /server
EXPOSE 8080
ENTRYPOINT ["/server"]
```

- Build stage: Chainguard `go` image with Go toolchain
- Runtime stage: Chainguard `static` image (~2MB, distroless)

## Dependencies

### Go Packages
- `github.com/labstack/echo/v4` - Web framework
- `github.com/labstack/echo/v4/middleware` - CORS, logging, recovery
- `github.com/google/uuid` - UUID generation (v4 and v7)
- `gopkg.in/yaml.v3` - YAML parsing/encoding
- `github.com/pelletier/go-toml/v2` - TOML parsing/encoding
- Standard library: `html/template`, `embed`, `encoding/json`, `time`

### Frontend (CDN)
- Pico CSS: `https://cdn.jsdelivr.net/npm/@picocss/pico@2/css/pico.min.css`
- HTMX: `https://unpkg.com/htmx.org@2`

### Syntax Highlighting
Custom CSS classes applied during JSON/YAML/TOML formatting. Go code wraps tokens in `<span class="string">`, `<span class="number">`, etc. Simple regex-based approach.

Copy-to-clipboard uses inline script with Clipboard API. No other JavaScript beyond HTMX.
