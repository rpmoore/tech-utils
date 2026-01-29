# Developer Tools Web Application Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a public utility website with Go/Echo/HTMX providing developer debugging tools (UUID generator, JSON formatter, format converter, timezone converter, unit converter).

**Architecture:** Echo serves HTML templates with embedded assets. HTMX handles form submissions via POST endpoints that return HTML fragments. Pico CSS provides styling via CDN. Configuration via environment variables.

**Tech Stack:** Go 1.25, Echo v4, HTMX 2.x, Pico CSS 2.x, Chainguard Docker images

**Testing:** Map-based table tests with `t.Cleanup()` for cleanup. Tests must be deterministic and parallelizable.

**Logging:** Use stdlib `slog.Logger` for structured logging.

---

## Task 1: Project Setup and Dependencies

**Files:**
- Modify: `go.mod`
- Create: `main.go`
- Create: `config/config.go`

**Step 1: Add dependencies to go.mod**

Run:
```bash
cd /Users/rmoore/code/tech-utils && go get github.com/labstack/echo/v4@latest github.com/labstack/echo/v4/middleware@latest github.com/google/uuid@latest gopkg.in/yaml.v3@latest github.com/pelletier/go-toml/v2@latest
```

**Step 2: Create config package**

Create `config/config.go`:
```go
package config

import (
	"log/slog"
	"os"
	"strings"
)

type Config struct {
	Port           string
	LogLevel       slog.Level
	AllowedOrigins []string
}

func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logLevelStr := os.Getenv("LOG_LEVEL")
	logLevel := parseLogLevel(logLevelStr)

	originsStr := os.Getenv("ALLOWED_ORIGINS")
	if originsStr == "" {
		originsStr = "*"
	}
	origins := strings.Split(originsStr, ",")
	for i := range origins {
		origins[i] = strings.TrimSpace(origins[i])
	}

	return &Config{
		Port:           port,
		LogLevel:       logLevel,
		AllowedOrigins: origins,
	}
}

func parseLogLevel(s string) slog.Level {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
```

**Step 3: Create minimal main.go**

Create `main.go`:
```go
package main

import (
	"log/slog"
	"os"

	"tech-utils/config"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	cfg := config.Load()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: cfg.LogLevel,
	}))
	slog.SetDefault(logger)

	e := echo.New()
	e.HideBanner = true

	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:   true,
		LogURI:      true,
		LogError:    true,
		HandleError: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			if v.Error == nil {
				logger.Info("request",
					slog.String("uri", v.URI),
					slog.Int("status", v.Status),
				)
			} else {
				logger.Error("request error",
					slog.String("uri", v.URI),
					slog.Int("status", v.Status),
					slog.String("error", v.Error.Error()),
				)
			}
			return nil
		},
	}))
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: cfg.AllowedOrigins,
	}))

	e.GET("/health", func(c echo.Context) error {
		return c.String(200, "OK")
	})

	logger.Info("starting server", slog.String("port", cfg.Port))
	e.Logger.Fatal(e.Start(":" + cfg.Port))
}
```

**Step 4: Verify it builds and runs**

Run:
```bash
cd /Users/rmoore/code/tech-utils && go build ./... && go run . &
sleep 2 && curl -s http://localhost:8080/health && pkill -f "tech-utils"
```
Expected: `OK`

**Step 5: Commit**

```bash
git add -A && git commit -m "feat: project setup with Echo and config"
```

---

## Task 2: Base Layout Template

**Files:**
- Create: `templates/layout.html`
- Create: `static/syntax.css`
- Modify: `main.go`

**Step 1: Create syntax.css**

Create `static/syntax.css`:
```css
.syntax-string { color: #22863a; }
.syntax-number { color: #005cc5; }
.syntax-boolean { color: #d73a49; }
.syntax-null { color: #6f42c1; }
.syntax-key { color: #032f62; }

.result-container {
    margin-top: 1rem;
}

.uuid-list {
    font-family: monospace;
    list-style: none;
    padding: 0;
}

.uuid-list li {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0.5rem;
    border-bottom: 1px solid var(--pico-muted-border-color);
}

.copy-btn {
    padding: 0.25rem 0.5rem;
    font-size: 0.875rem;
}

pre.output {
    background: var(--pico-code-background-color);
    padding: 1rem;
    border-radius: 0.25rem;
    overflow-x: auto;
    white-space: pre-wrap;
    word-wrap: break-word;
}

.error {
    color: var(--pico-del-color);
    padding: 1rem;
    border: 1px solid var(--pico-del-color);
    border-radius: 0.25rem;
}

.formula {
    font-style: italic;
    color: var(--pico-muted-color);
    margin-top: 0.5rem;
}
```

**Step 2: Create layout template**

Create `templates/layout.html`:
```html
<!DOCTYPE html>
<html lang="en" data-theme="light">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}} - Dev Tools</title>
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/@picocss/pico@2/css/pico.min.css">
    <link rel="stylesheet" href="/static/syntax.css">
    <script src="https://unpkg.com/htmx.org@2"></script>
</head>
<body>
    <nav class="container">
        <ul>
            <li><strong><a href="/">Dev Tools</a></strong></li>
        </ul>
        <ul>
            <li><a href="/uuid">UUID</a></li>
            <li><a href="/json">JSON</a></li>
            <li><a href="/convert">Convert</a></li>
            <li><a href="/timezone">Timezone</a></li>
            <li><a href="/units">Units</a></li>
        </ul>
    </nav>
    <main class="container">
        {{template "content" .}}
    </main>
    <footer class="container">
        <small>Developer utility tools. No data stored.</small>
    </footer>
    <script>
        function copyToClipboard(text, btn) {
            navigator.clipboard.writeText(text).then(function() {
                var original = btn.textContent;
                btn.textContent = 'Copied!';
                setTimeout(function() { btn.textContent = original; }, 1500);
            });
        }
    </script>
</body>
</html>
```

**Step 3: Update main.go to serve static files and set up template rendering**

Replace `main.go` with:
```go
package main

import (
	"embed"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"os"

	"tech-utils/config"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

//go:embed templates/*.html
var templatesFS embed.FS

//go:embed static/*
var staticFS embed.FS

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func main() {
	cfg := config.Load()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: cfg.LogLevel,
	}))
	slog.SetDefault(logger)

	e := echo.New()
	e.HideBanner = true

	tmpl := template.Must(template.ParseFS(templatesFS, "templates/*.html"))
	e.Renderer = &Template{templates: tmpl}

	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:   true,
		LogURI:      true,
		LogError:    true,
		HandleError: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			if v.Error == nil {
				logger.Info("request",
					slog.String("uri", v.URI),
					slog.Int("status", v.Status),
				)
			} else {
				logger.Error("request error",
					slog.String("uri", v.URI),
					slog.Int("status", v.Status),
					slog.String("error", v.Error.Error()),
				)
			}
			return nil
		},
	}))
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: cfg.AllowedOrigins,
	}))

	e.GET("/static/*", echo.WrapHandler(http.FileServer(http.FS(staticFS))))

	e.GET("/health", func(c echo.Context) error {
		return c.String(200, "OK")
	})

	logger.Info("starting server", slog.String("port", cfg.Port))
	e.Logger.Fatal(e.Start(":" + cfg.Port))
}
```

**Step 4: Verify build**

Run:
```bash
cd /Users/rmoore/code/tech-utils && go build ./...
```
Expected: No errors

**Step 5: Commit**

```bash
git add -A && git commit -m "feat: add base layout template and static assets"
```

---

## Task 3: Home Page

**Files:**
- Create: `handlers/home.go`
- Create: `handlers/home_test.go`
- Create: `templates/home.html`
- Modify: `main.go`

**Step 1: Write the failing test**

Create `handlers/home_test.go`:
```go
package handlers

import (
	"testing"
)

func TestHomeHandler_GetPageData(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		wantTitle     string
		wantToolCount int
		wantTools     []string
	}{
		"returns correct page data": {
			wantTitle:     "Home",
			wantToolCount: 5,
			wantTools:     []string{"UUID Generator", "JSON Formatter", "Format Converter", "Timezone Converter", "Unit Converter"},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			h := &HomeHandler{}
			data := h.GetPageData()

			if data["Title"] != tt.wantTitle {
				t.Errorf("Title = %v, want %v", data["Title"], tt.wantTitle)
			}

			tools, ok := data["Tools"].([]Tool)
			if !ok {
				t.Fatal("Tools is not []Tool")
			}

			if len(tools) != tt.wantToolCount {
				t.Errorf("len(Tools) = %d, want %d", len(tools), tt.wantToolCount)
			}

			for i, tool := range tools {
				if tool.Name != tt.wantTools[i] {
					t.Errorf("Tools[%d].Name = %s, want %s", i, tool.Name, tt.wantTools[i])
				}
			}
		})
	}
}
```

**Step 2: Run test to verify it fails**

Run:
```bash
cd /Users/rmoore/code/tech-utils && go test ./handlers/... -v
```
Expected: FAIL (package doesn't exist yet)

**Step 3: Create home handler**

Create `handlers/home.go`:
```go
package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type Tool struct {
	Name        string
	Description string
	Path        string
}

type HomeHandler struct{}

func NewHomeHandler() *HomeHandler {
	return &HomeHandler{}
}

func (h *HomeHandler) GetPageData() map[string]interface{} {
	tools := []Tool{
		{Name: "UUID Generator", Description: "Generate UUIDs (v4 and v7)", Path: "/uuid"},
		{Name: "JSON Formatter", Description: "Pretty-print and syntax highlight JSON", Path: "/json"},
		{Name: "Format Converter", Description: "Convert between JSON, YAML, and TOML", Path: "/convert"},
		{Name: "Timezone Converter", Description: "Convert times between timezones", Path: "/timezone"},
		{Name: "Unit Converter", Description: "Convert between units of measurement", Path: "/units"},
	}

	return map[string]interface{}{
		"Title": "Home",
		"Tools": tools,
	}
}

func (h *HomeHandler) Index(c echo.Context) error {
	return c.Render(http.StatusOK, "home.html", h.GetPageData())
}
```

**Step 4: Run test to verify it passes**

Run:
```bash
cd /Users/rmoore/code/tech-utils && go test ./handlers/... -v
```
Expected: PASS

**Step 5: Create home template**

Create `templates/home.html`:
```html
{{define "content"}}
<h1>Developer Tools</h1>
<p>Useful utilities for debugging and development.</p>

<div class="grid">
    {{range .Tools}}
    <article>
        <header>{{.Name}}</header>
        <p>{{.Description}}</p>
        <footer>
            <a href="{{.Path}}" role="button">Open</a>
        </footer>
    </article>
    {{end}}
</div>
{{end}}
```

**Step 6: Register route in main.go**

Add import and route to `main.go`. After the static files handler, add:
```go
// Import handlers
import "tech-utils/handlers"

// In main(), after static files:
homeHandler := handlers.NewHomeHandler()
e.GET("/", homeHandler.Index)
```

**Step 7: Verify build and manual test**

Run:
```bash
cd /Users/rmoore/code/tech-utils && go build ./... && go run . &
sleep 2 && curl -s http://localhost:8080/ | grep -o "Developer Tools" && pkill -f "tech-utils"
```
Expected: `Developer Tools`

**Step 8: Commit**

```bash
git add -A && git commit -m "feat: add home page with tool cards"
```

---

## Task 4: UUID Generator

**Files:**
- Create: `handlers/uuid.go`
- Create: `handlers/uuid_test.go`
- Create: `templates/uuid.html`
- Modify: `main.go`

**Step 1: Write failing tests**

Create `handlers/uuid_test.go`:
```go
package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestUUIDHandler_Generate(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		count      string
		version    string
		wantStatus int
		wantInBody []string
		wantCount  int
	}{
		"v4 single uuid": {
			count:      "1",
			version:    "4",
			wantStatus: http.StatusOK,
			wantInBody: []string{"-", "<li>"},
			wantCount:  1,
		},
		"v7 multiple uuids": {
			count:      "5",
			version:    "7",
			wantStatus: http.StatusOK,
			wantInBody: []string{"<li>"},
			wantCount:  5,
		},
		"invalid count over limit": {
			count:      "150",
			version:    "4",
			wantStatus: http.StatusOK,
			wantInBody: []string{"error", "1-100"},
			wantCount:  0,
		},
		"invalid count zero": {
			count:      "0",
			version:    "4",
			wantStatus: http.StatusOK,
			wantInBody: []string{"error"},
			wantCount:  0,
		},
		"invalid count not a number": {
			count:      "abc",
			version:    "4",
			wantStatus: http.StatusOK,
			wantInBody: []string{"error"},
			wantCount:  0,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			e := echo.New()
			form := url.Values{}
			form.Add("count", tt.count)
			form.Add("version", tt.version)

			req := httptest.NewRequest(http.MethodPost, "/uuid/generate", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.Header.Set("HX-Request", "true")
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			h := NewUUIDHandler()
			if err := h.Generate(c); err != nil {
				t.Fatalf("handler returned error: %v", err)
			}

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}

			body := rec.Body.String()
			for _, want := range tt.wantInBody {
				if !strings.Contains(body, want) {
					t.Errorf("body missing %q, got: %s", want, body)
				}
			}

			if tt.wantCount > 0 {
				count := strings.Count(body, "<li>")
				if count != tt.wantCount {
					t.Errorf("uuid count = %d, want %d", count, tt.wantCount)
				}
			}
		})
	}
}
```

**Step 2: Run test to verify it fails**

Run:
```bash
cd /Users/rmoore/code/tech-utils && go test ./handlers/... -v -run UUID
```
Expected: FAIL

**Step 3: Implement UUID handler**

Create `handlers/uuid.go`:
```go
package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type UUIDHandler struct{}

func NewUUIDHandler() *UUIDHandler {
	return &UUIDHandler{}
}

func (h *UUIDHandler) Index(c echo.Context) error {
	return c.Render(http.StatusOK, "uuid.html", map[string]interface{}{
		"Title": "UUID Generator",
	})
}

func (h *UUIDHandler) Generate(c echo.Context) error {
	countStr := c.FormValue("count")
	version := c.FormValue("version")

	count, err := strconv.Atoi(countStr)
	if err != nil || count < 1 || count > 100 {
		return c.HTML(http.StatusOK, `<div class="error">Count must be between 1-100</div>`)
	}

	var uuids []string
	for i := 0; i < count; i++ {
		var u uuid.UUID
		if version == "7" {
			u, err = uuid.NewV7()
		} else {
			u, err = uuid.NewRandom()
		}
		if err != nil {
			return c.HTML(http.StatusOK, fmt.Sprintf(`<div class="error">Error generating UUID: %s</div>`, template.HTMLEscapeString(err.Error())))
		}
		uuids = append(uuids, u.String())
	}

	var sb strings.Builder
	sb.WriteString(`<ul class="uuid-list">`)
	for _, u := range uuids {
		sb.WriteString(fmt.Sprintf(`<li><code>%s</code><button class="copy-btn secondary outline" onclick="copyToClipboard('%s', this)">Copy</button></li>`, u, u))
	}
	sb.WriteString(`</ul>`)

	return c.HTML(http.StatusOK, sb.String())
}
```

**Step 4: Run test to verify it passes**

Run:
```bash
cd /Users/rmoore/code/tech-utils && go test ./handlers/... -v -run UUID
```
Expected: PASS

**Step 5: Create UUID template**

Create `templates/uuid.html`:
```html
{{define "content"}}
<h1>UUID Generator</h1>
<p>Generate random UUIDs (v4) or time-ordered UUIDs (v7).</p>

<form hx-post="/uuid/generate" hx-target="#result" hx-swap="innerHTML">
    <div class="grid">
        <label>
            Count (1-100)
            <input type="number" name="count" value="1" min="1" max="100" required>
        </label>
        <label>
            Version
            <select name="version">
                <option value="4">v4 (Random)</option>
                <option value="7">v7 (Time-ordered)</option>
            </select>
        </label>
    </div>
    <button type="submit">Generate</button>
</form>

<div id="result" class="result-container"></div>
{{end}}
```

**Step 6: Register routes in main.go**

Add to `main.go` after home route:
```go
uuidHandler := handlers.NewUUIDHandler()
e.GET("/uuid", uuidHandler.Index)
e.POST("/uuid/generate", uuidHandler.Generate)
```

**Step 7: Commit**

```bash
git add -A && git commit -m "feat: add UUID generator tool"
```

---

## Task 5: JSON Formatter

**Files:**
- Create: `handlers/json.go`
- Create: `handlers/json_test.go`
- Create: `templates/json.html`
- Modify: `main.go`

**Step 1: Write failing tests**

Create `handlers/json_test.go`:
```go
package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestJSONHandler_Format(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input      string
		indent     string
		wantStatus int
		wantInBody []string
	}{
		"valid json with syntax highlighting": {
			input:      `{"name":"test","value":123}`,
			indent:     "2",
			wantStatus: http.StatusOK,
			wantInBody: []string{"syntax-string", "syntax-number"},
		},
		"invalid json returns error": {
			input:      `{invalid json}`,
			indent:     "2",
			wantStatus: http.StatusOK,
			wantInBody: []string{"error"},
		},
		"tab indent": {
			input:      `{"a":1}`,
			indent:     "tab",
			wantStatus: http.StatusOK,
			wantInBody: []string{"<pre"},
		},
		"4 space indent": {
			input:      `{"a":1}`,
			indent:     "4",
			wantStatus: http.StatusOK,
			wantInBody: []string{"<pre"},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			e := echo.New()
			form := url.Values{}
			form.Add("input", tt.input)
			form.Add("indent", tt.indent)

			req := httptest.NewRequest(http.MethodPost, "/json/format", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.Header.Set("HX-Request", "true")
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			h := NewJSONHandler()
			if err := h.Format(c); err != nil {
				t.Fatalf("handler returned error: %v", err)
			}

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}

			body := rec.Body.String()
			for _, want := range tt.wantInBody {
				if !strings.Contains(body, want) {
					t.Errorf("body missing %q, got: %s", want, body)
				}
			}
		})
	}
}
```

**Step 2: Run test to verify it fails**

Run:
```bash
cd /Users/rmoore/code/tech-utils && go test ./handlers/... -v -run JSON
```
Expected: FAIL

**Step 3: Implement JSON handler**

Create `handlers/json.go`:
```go
package handlers

import (
	"bytes"
	"encoding/json"
	"html"
	"net/http"
	"regexp"
	"strings"

	"github.com/labstack/echo/v4"
)

type JSONHandler struct{}

func NewJSONHandler() *JSONHandler {
	return &JSONHandler{}
}

func (h *JSONHandler) Index(c echo.Context) error {
	return c.Render(http.StatusOK, "json.html", map[string]interface{}{
		"Title": "JSON Formatter",
	})
}

func (h *JSONHandler) Format(c echo.Context) error {
	input := c.FormValue("input")
	indentType := c.FormValue("indent")

	var parsed interface{}
	if err := json.Unmarshal([]byte(input), &parsed); err != nil {
		return c.HTML(http.StatusOK, `<div class="error">Invalid JSON: `+html.EscapeString(err.Error())+`</div>`)
	}

	var indent string
	switch indentType {
	case "tab":
		indent = "\t"
	case "4":
		indent = "    "
	default:
		indent = "  "
	}

	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetIndent("", indent)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(parsed); err != nil {
		return c.HTML(http.StatusOK, `<div class="error">Error formatting: `+html.EscapeString(err.Error())+`</div>`)
	}

	formatted := strings.TrimSuffix(buf.String(), "\n")
	highlighted := syntaxHighlightJSON(formatted)

	return c.HTML(http.StatusOK, `<pre class="output">`+highlighted+`</pre>`)
}

func syntaxHighlightJSON(s string) string {
	s = html.EscapeString(s)

	stringRe := regexp.MustCompile(`&quot;([^&]*)&quot;`)
	s = stringRe.ReplaceAllString(s, `<span class="syntax-string">&quot;$1&quot;</span>`)

	numberRe := regexp.MustCompile(`\b(-?\d+\.?\d*)\b`)
	s = numberRe.ReplaceAllString(s, `<span class="syntax-number">$1</span>`)

	boolRe := regexp.MustCompile(`\b(true|false)\b`)
	s = boolRe.ReplaceAllString(s, `<span class="syntax-boolean">$1</span>`)

	nullRe := regexp.MustCompile(`\bnull\b`)
	s = nullRe.ReplaceAllString(s, `<span class="syntax-null">null</span>`)

	return s
}
```

**Step 4: Run test to verify it passes**

Run:
```bash
cd /Users/rmoore/code/tech-utils && go test ./handlers/... -v -run JSON
```
Expected: PASS

**Step 5: Create JSON template**

Create `templates/json.html`:
```html
{{define "content"}}
<h1>JSON Formatter</h1>
<p>Pretty-print and syntax highlight JSON.</p>

<form hx-post="/json/format" hx-target="#result" hx-swap="innerHTML">
    <label>
        JSON Input
        <textarea name="input" rows="10" placeholder='{"key": "value"}'></textarea>
    </label>
    <label>
        Indent
        <select name="indent">
            <option value="2">2 spaces</option>
            <option value="4">4 spaces</option>
            <option value="tab">Tabs</option>
        </select>
    </label>
    <button type="submit">Format</button>
</form>

<div id="result" class="result-container"></div>
{{end}}
```

**Step 6: Register routes in main.go**

Add to `main.go`:
```go
jsonHandler := handlers.NewJSONHandler()
e.GET("/json", jsonHandler.Index)
e.POST("/json/format", jsonHandler.Format)
```

**Step 7: Commit**

```bash
git add -A && git commit -m "feat: add JSON formatter with syntax highlighting"
```

---

## Task 6: Format Converter (JSON/YAML/TOML)

**Files:**
- Create: `handlers/convert.go`
- Create: `handlers/convert_test.go`
- Create: `templates/convert.html`
- Modify: `main.go`

**Step 1: Write failing tests**

Create `handlers/convert_test.go`:
```go
package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestConvertHandler_Transform(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input      string
		from       string
		to         string
		wantStatus int
		wantInBody []string
	}{
		"json to yaml": {
			input:      `{"name":"test","count":42}`,
			from:       "json",
			to:         "yaml",
			wantStatus: http.StatusOK,
			wantInBody: []string{"name:", "test"},
		},
		"yaml to json": {
			input:      "name: test\ncount: 42",
			from:       "yaml",
			to:         "json",
			wantStatus: http.StatusOK,
			wantInBody: []string{`"name"`, `"test"`},
		},
		"json to toml": {
			input:      `{"name":"test","count":42}`,
			from:       "json",
			to:         "toml",
			wantStatus: http.StatusOK,
			wantInBody: []string{"name", "="},
		},
		"invalid json input": {
			input:      `{invalid}`,
			from:       "json",
			to:         "yaml",
			wantStatus: http.StatusOK,
			wantInBody: []string{"error"},
		},
		"unknown source format": {
			input:      `test`,
			from:       "xml",
			to:         "json",
			wantStatus: http.StatusOK,
			wantInBody: []string{"error", "Unknown source format"},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			e := echo.New()
			form := url.Values{}
			form.Add("input", tt.input)
			form.Add("from", tt.from)
			form.Add("to", tt.to)

			req := httptest.NewRequest(http.MethodPost, "/convert/transform", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			h := NewConvertHandler()
			if err := h.Transform(c); err != nil {
				t.Fatalf("handler returned error: %v", err)
			}

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}

			body := rec.Body.String()
			for _, want := range tt.wantInBody {
				if !strings.Contains(body, want) {
					t.Errorf("body missing %q, got: %s", want, body)
				}
			}
		})
	}
}
```

**Step 2: Run test to verify it fails**

Run:
```bash
cd /Users/rmoore/code/tech-utils && go test ./handlers/... -v -run Convert
```
Expected: FAIL

**Step 3: Implement convert handler**

Create `handlers/convert.go`:
```go
package handlers

import (
	"bytes"
	"encoding/json"
	"html"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
)

type ConvertHandler struct{}

func NewConvertHandler() *ConvertHandler {
	return &ConvertHandler{}
}

func (h *ConvertHandler) Index(c echo.Context) error {
	return c.Render(http.StatusOK, "convert.html", map[string]interface{}{
		"Title": "Format Converter",
	})
}

func (h *ConvertHandler) Transform(c echo.Context) error {
	input := c.FormValue("input")
	fromFormat := c.FormValue("from")
	toFormat := c.FormValue("to")

	var data interface{}
	var err error

	switch fromFormat {
	case "json":
		err = json.Unmarshal([]byte(input), &data)
	case "yaml":
		err = yaml.Unmarshal([]byte(input), &data)
	case "toml":
		err = toml.Unmarshal([]byte(input), &data)
	default:
		return c.HTML(http.StatusOK, `<div class="error">Unknown source format</div>`)
	}

	if err != nil {
		return c.HTML(http.StatusOK, `<div class="error">Error parsing `+fromFormat+`: `+html.EscapeString(err.Error())+`</div>`)
	}

	var output []byte

	switch toFormat {
	case "json":
		var buf bytes.Buffer
		encoder := json.NewEncoder(&buf)
		encoder.SetIndent("", "  ")
		encoder.SetEscapeHTML(false)
		err = encoder.Encode(data)
		output = bytes.TrimSuffix(buf.Bytes(), []byte("\n"))
	case "yaml":
		output, err = yaml.Marshal(data)
	case "toml":
		output, err = toml.Marshal(data)
	default:
		return c.HTML(http.StatusOK, `<div class="error">Unknown target format</div>`)
	}

	if err != nil {
		return c.HTML(http.StatusOK, `<div class="error">Error converting to `+toFormat+`: `+html.EscapeString(err.Error())+`</div>`)
	}

	highlighted := html.EscapeString(string(output))
	if toFormat == "json" {
		highlighted = syntaxHighlightJSON(string(output))
	}

	return c.HTML(http.StatusOK, `<pre class="output">`+highlighted+`</pre>`)
}
```

**Step 4: Run test to verify it passes**

Run:
```bash
cd /Users/rmoore/code/tech-utils && go test ./handlers/... -v -run Convert
```
Expected: PASS

**Step 5: Create convert template**

Create `templates/convert.html`:
```html
{{define "content"}}
<h1>Format Converter</h1>
<p>Convert between JSON, YAML, and TOML formats.</p>

<form hx-post="/convert/transform" hx-target="#result" hx-swap="innerHTML">
    <label>
        Input
        <textarea name="input" rows="10" placeholder="Paste JSON, YAML, or TOML here"></textarea>
    </label>
    <div class="grid">
        <label>
            From
            <select name="from">
                <option value="json">JSON</option>
                <option value="yaml">YAML</option>
                <option value="toml">TOML</option>
            </select>
        </label>
        <label>
            To
            <select name="to">
                <option value="yaml">YAML</option>
                <option value="json">JSON</option>
                <option value="toml">TOML</option>
            </select>
        </label>
    </div>
    <button type="submit">Convert</button>
</form>

<div id="result" class="result-container"></div>
{{end}}
```

**Step 6: Register routes in main.go**

Add to `main.go`:
```go
convertHandler := handlers.NewConvertHandler()
e.GET("/convert", convertHandler.Index)
e.POST("/convert/transform", convertHandler.Transform)
```

**Step 7: Commit**

```bash
git add -A && git commit -m "feat: add format converter for JSON/YAML/TOML"
```

---

## Task 7: Timezone Converter

**Files:**
- Create: `handlers/timezone.go`
- Create: `handlers/timezone_test.go`
- Create: `templates/timezone.html`
- Modify: `main.go`

**Step 1: Write failing tests**

Create `handlers/timezone_test.go`:
```go
package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestTimezoneHandler_Convert(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		datetime   string
		from       string
		to         string
		wantStatus int
		wantInBody []string
	}{
		"new york to london summer": {
			datetime:   "2024-06-15T10:30:00",
			from:       "America/New_York",
			to:         "Europe/London",
			wantStatus: http.StatusOK,
			wantInBody: []string{"15:30"},
		},
		"utc to utc": {
			datetime:   "2024-01-01T12:00:00",
			from:       "UTC",
			to:         "UTC",
			wantStatus: http.StatusOK,
			wantInBody: []string{"12:00"},
		},
		"invalid datetime": {
			datetime:   "not-a-date",
			from:       "UTC",
			to:         "UTC",
			wantStatus: http.StatusOK,
			wantInBody: []string{"error", "Invalid datetime"},
		},
		"invalid source timezone": {
			datetime:   "2024-06-15T10:30:00",
			from:       "Invalid/Zone",
			to:         "UTC",
			wantStatus: http.StatusOK,
			wantInBody: []string{"error", "Invalid source timezone"},
		},
		"invalid target timezone": {
			datetime:   "2024-06-15T10:30:00",
			from:       "UTC",
			to:         "Invalid/Zone",
			wantStatus: http.StatusOK,
			wantInBody: []string{"error", "Invalid target timezone"},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			e := echo.New()
			form := url.Values{}
			form.Add("datetime", tt.datetime)
			form.Add("from", tt.from)
			form.Add("to", tt.to)

			req := httptest.NewRequest(http.MethodPost, "/timezone/convert", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			h := NewTimezoneHandler()
			if err := h.Convert(c); err != nil {
				t.Fatalf("handler returned error: %v", err)
			}

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}

			body := rec.Body.String()
			for _, want := range tt.wantInBody {
				if !strings.Contains(body, want) {
					t.Errorf("body missing %q, got: %s", want, body)
				}
			}
		})
	}
}
```

**Step 2: Run test to verify it fails**

Run:
```bash
cd /Users/rmoore/code/tech-utils && go test ./handlers/... -v -run Timezone
```
Expected: FAIL

**Step 3: Implement timezone handler**

Create `handlers/timezone.go`:
```go
package handlers

import (
	"fmt"
	"html"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

var CommonTimezones = []string{
	"UTC",
	"America/New_York",
	"America/Chicago",
	"America/Denver",
	"America/Los_Angeles",
	"America/Toronto",
	"America/Vancouver",
	"Europe/London",
	"Europe/Paris",
	"Europe/Berlin",
	"Europe/Moscow",
	"Asia/Tokyo",
	"Asia/Shanghai",
	"Asia/Singapore",
	"Asia/Dubai",
	"Asia/Kolkata",
	"Australia/Sydney",
	"Australia/Melbourne",
	"Pacific/Auckland",
}

type TimezoneHandler struct{}

func NewTimezoneHandler() *TimezoneHandler {
	return &TimezoneHandler{}
}

func (h *TimezoneHandler) Index(c echo.Context) error {
	return c.Render(http.StatusOK, "timezone.html", map[string]interface{}{
		"Title":     "Timezone Converter",
		"Timezones": CommonTimezones,
	})
}

func (h *TimezoneHandler) Convert(c echo.Context) error {
	datetime := c.FormValue("datetime")
	fromTZ := c.FormValue("from")
	toTZ := c.FormValue("to")

	fromLoc, err := time.LoadLocation(fromTZ)
	if err != nil {
		return c.HTML(http.StatusOK, `<div class="error">Invalid source timezone: `+html.EscapeString(err.Error())+`</div>`)
	}

	toLoc, err := time.LoadLocation(toTZ)
	if err != nil {
		return c.HTML(http.StatusOK, `<div class="error">Invalid target timezone: `+html.EscapeString(err.Error())+`</div>`)
	}

	formats := []string{
		"2006-01-02T15:04:05",
		"2006-01-02T15:04",
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02",
	}

	var t time.Time
	for _, format := range formats {
		t, err = time.ParseInLocation(format, datetime, fromLoc)
		if err == nil {
			break
		}
	}

	if err != nil {
		return c.HTML(http.StatusOK, `<div class="error">Invalid datetime format. Use: YYYY-MM-DD HH:MM:SS or YYYY-MM-DDTHH:MM:SS</div>`)
	}

	converted := t.In(toLoc)

	result := fmt.Sprintf(`
<article>
    <header>Converted Time</header>
    <dl>
        <dt>ISO 8601</dt>
        <dd><code>%s</code></dd>
        <dt>RFC 3339</dt>
        <dd><code>%s</code></dd>
        <dt>Unix Timestamp</dt>
        <dd><code>%d</code></dd>
        <dt>Human Readable</dt>
        <dd><code>%s</code></dd>
    </dl>
</article>`,
		converted.Format("2006-01-02T15:04:05-07:00"),
		converted.Format(time.RFC3339),
		converted.Unix(),
		converted.Format("Monday, January 2, 2006 at 3:04 PM MST"),
	)

	return c.HTML(http.StatusOK, result)
}
```

**Step 4: Run test to verify it passes**

Run:
```bash
cd /Users/rmoore/code/tech-utils && go test ./handlers/... -v -run Timezone
```
Expected: PASS

**Step 5: Create timezone template**

Create `templates/timezone.html`:
```html
{{define "content"}}
<h1>Timezone Converter</h1>
<p>Convert times between different timezones.</p>

<form hx-post="/timezone/convert" hx-target="#result" hx-swap="innerHTML">
    <label>
        Date/Time
        <input type="datetime-local" name="datetime" required>
    </label>
    <div class="grid">
        <label>
            From Timezone
            <select name="from">
                {{range .Timezones}}
                <option value="{{.}}" {{if eq . "UTC"}}selected{{end}}>{{.}}</option>
                {{end}}
            </select>
        </label>
        <label>
            To Timezone
            <select name="to">
                {{range .Timezones}}
                <option value="{{.}}" {{if eq . "America/New_York"}}selected{{end}}>{{.}}</option>
                {{end}}
            </select>
        </label>
    </div>
    <button type="submit">Convert</button>
</form>

<div id="result" class="result-container"></div>
{{end}}
```

**Step 6: Register routes in main.go**

Add to `main.go`:
```go
timezoneHandler := handlers.NewTimezoneHandler()
e.GET("/timezone", timezoneHandler.Index)
e.POST("/timezone/convert", timezoneHandler.Convert)
```

**Step 7: Commit**

```bash
git add -A && git commit -m "feat: add timezone converter tool"
```

---

## Task 8: Unit Converter

**Files:**
- Create: `handlers/units.go`
- Create: `handlers/units_test.go`
- Create: `templates/units.html`
- Modify: `main.go`

**Step 1: Write failing tests**

Create `handlers/units_test.go`:
```go
package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestUnitsHandler_Convert(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		value      string
		category   string
		from       string
		to         string
		wantStatus int
		wantInBody []string
	}{
		"meters to feet": {
			value:      "1",
			category:   "length",
			from:       "m",
			to:         "ft",
			wantStatus: http.StatusOK,
			wantInBody: []string{"3.28"},
		},
		"celsius to fahrenheit": {
			value:      "100",
			category:   "temperature",
			from:       "c",
			to:         "f",
			wantStatus: http.StatusOK,
			wantInBody: []string{"212"},
		},
		"gigabytes to megabytes": {
			value:      "1",
			category:   "data",
			from:       "gb",
			to:         "mb",
			wantStatus: http.StatusOK,
			wantInBody: []string{"1024"},
		},
		"kilograms to pounds": {
			value:      "1",
			category:   "weight",
			from:       "kg",
			to:         "lb",
			wantStatus: http.StatusOK,
			wantInBody: []string{"2.2"},
		},
		"invalid value": {
			value:      "not-a-number",
			category:   "length",
			from:       "m",
			to:         "ft",
			wantStatus: http.StatusOK,
			wantInBody: []string{"error"},
		},
		"unknown category": {
			value:      "1",
			category:   "unknown",
			from:       "m",
			to:         "ft",
			wantStatus: http.StatusOK,
			wantInBody: []string{"error", "Unknown category"},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			e := echo.New()
			form := url.Values{}
			form.Add("value", tt.value)
			form.Add("category", tt.category)
			form.Add("from", tt.from)
			form.Add("to", tt.to)

			req := httptest.NewRequest(http.MethodPost, "/units/convert", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			h := NewUnitsHandler()
			if err := h.Convert(c); err != nil {
				t.Fatalf("handler returned error: %v", err)
			}

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}

			body := rec.Body.String()
			for _, want := range tt.wantInBody {
				if !strings.Contains(body, want) {
					t.Errorf("body missing %q, got: %s", want, body)
				}
			}
		})
	}
}
```

**Step 2: Run test to verify it fails**

Run:
```bash
cd /Users/rmoore/code/tech-utils && go test ./handlers/... -v -run Units
```
Expected: FAIL

**Step 3: Implement units handler**

Create `handlers/units.go`:
```go
package handlers

import (
	"fmt"
	"html"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

type UnitCategory struct {
	Name  string
	Units []Unit
}

type Unit struct {
	Code string
	Name string
}

var UnitCategories = []UnitCategory{
	{
		Name: "length",
		Units: []Unit{
			{Code: "m", Name: "Meters"},
			{Code: "km", Name: "Kilometers"},
			{Code: "cm", Name: "Centimeters"},
			{Code: "mm", Name: "Millimeters"},
			{Code: "mi", Name: "Miles"},
			{Code: "yd", Name: "Yards"},
			{Code: "ft", Name: "Feet"},
			{Code: "in", Name: "Inches"},
		},
	},
	{
		Name: "weight",
		Units: []Unit{
			{Code: "kg", Name: "Kilograms"},
			{Code: "g", Name: "Grams"},
			{Code: "mg", Name: "Milligrams"},
			{Code: "lb", Name: "Pounds"},
			{Code: "oz", Name: "Ounces"},
		},
	},
	{
		Name: "temperature",
		Units: []Unit{
			{Code: "c", Name: "Celsius"},
			{Code: "f", Name: "Fahrenheit"},
			{Code: "k", Name: "Kelvin"},
		},
	},
	{
		Name: "data",
		Units: []Unit{
			{Code: "b", Name: "Bytes"},
			{Code: "kb", Name: "Kilobytes"},
			{Code: "mb", Name: "Megabytes"},
			{Code: "gb", Name: "Gigabytes"},
			{Code: "tb", Name: "Terabytes"},
		},
	},
}

var lengthToBase = map[string]float64{
	"m": 1, "km": 1000, "cm": 0.01, "mm": 0.001,
	"mi": 1609.344, "yd": 0.9144, "ft": 0.3048, "in": 0.0254,
}

var weightToBase = map[string]float64{
	"kg": 1000, "g": 1, "mg": 0.001, "lb": 453.592, "oz": 28.3495,
}

var dataToBase = map[string]float64{
	"b": 1, "kb": 1024, "mb": 1024 * 1024,
	"gb": 1024 * 1024 * 1024, "tb": 1024 * 1024 * 1024 * 1024,
}

type UnitsHandler struct{}

func NewUnitsHandler() *UnitsHandler {
	return &UnitsHandler{}
}

func (h *UnitsHandler) Index(c echo.Context) error {
	return c.Render(http.StatusOK, "units.html", map[string]interface{}{
		"Title":      "Unit Converter",
		"Categories": UnitCategories,
	})
}

func (h *UnitsHandler) Convert(c echo.Context) error {
	valueStr := c.FormValue("value")
	category := c.FormValue("category")
	from := c.FormValue("from")
	to := c.FormValue("to")

	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return c.HTML(http.StatusOK, `<div class="error">Invalid number: `+html.EscapeString(valueStr)+`</div>`)
	}

	var result float64
	var formula string

	switch category {
	case "length":
		fromFactor, ok1 := lengthToBase[from]
		toFactor, ok2 := lengthToBase[to]
		if !ok1 || !ok2 {
			return c.HTML(http.StatusOK, `<div class="error">Unknown length unit</div>`)
		}
		result = value * fromFactor / toFactor
		formula = fmt.Sprintf("%.6g %s * %.6g / %.6g = %.6g %s", value, from, fromFactor, toFactor, result, to)

	case "weight":
		fromFactor, ok1 := weightToBase[from]
		toFactor, ok2 := weightToBase[to]
		if !ok1 || !ok2 {
			return c.HTML(http.StatusOK, `<div class="error">Unknown weight unit</div>`)
		}
		result = value * fromFactor / toFactor
		formula = fmt.Sprintf("%.6g %s * %.6g / %.6g = %.6g %s", value, from, fromFactor, toFactor, result, to)

	case "temperature":
		result, formula = convertTemperature(value, from, to)
		if formula == "" {
			return c.HTML(http.StatusOK, `<div class="error">Unknown temperature unit</div>`)
		}

	case "data":
		fromFactor, ok1 := dataToBase[from]
		toFactor, ok2 := dataToBase[to]
		if !ok1 || !ok2 {
			return c.HTML(http.StatusOK, `<div class="error">Unknown data unit</div>`)
		}
		result = value * fromFactor / toFactor
		formula = fmt.Sprintf("%.6g %s * %.0f / %.0f = %.6g %s", value, from, fromFactor, toFactor, result, to)

	default:
		return c.HTML(http.StatusOK, `<div class="error">Unknown category</div>`)
	}

	output := fmt.Sprintf(`
<article>
    <header>Result</header>
    <p><strong>%.6g %s = %.6g %s</strong></p>
    <p class="formula">Formula: %s</p>
</article>`, value, from, result, to, formula)

	return c.HTML(http.StatusOK, output)
}

func convertTemperature(value float64, from, to string) (float64, string) {
	var celsius float64
	switch from {
	case "c":
		celsius = value
	case "f":
		celsius = (value - 32) * 5 / 9
	case "k":
		celsius = value - 273.15
	default:
		return 0, ""
	}

	var result float64
	var formula string
	switch to {
	case "c":
		result = celsius
		if from == "f" {
			formula = fmt.Sprintf("(%.6g - 32) * 5/9 = %.6g", value, result)
		} else if from == "k" {
			formula = fmt.Sprintf("%.6g - 273.15 = %.6g", value, result)
		} else {
			formula = fmt.Sprintf("%.6g C = %.6g C", value, result)
		}
	case "f":
		result = celsius*9/5 + 32
		if from == "c" {
			formula = fmt.Sprintf("%.6g * 9/5 + 32 = %.6g", value, result)
		} else if from == "k" {
			formula = fmt.Sprintf("(%.6g - 273.15) * 9/5 + 32 = %.6g", value, result)
		} else {
			formula = fmt.Sprintf("%.6g F = %.6g F", value, result)
		}
	case "k":
		result = celsius + 273.15
		if from == "c" {
			formula = fmt.Sprintf("%.6g + 273.15 = %.6g", value, result)
		} else if from == "f" {
			formula = fmt.Sprintf("(%.6g - 32) * 5/9 + 273.15 = %.6g", value, result)
		} else {
			formula = fmt.Sprintf("%.6g K = %.6g K", value, result)
		}
	default:
		return 0, ""
	}

	return result, formula
}
```

**Step 4: Run test to verify it passes**

Run:
```bash
cd /Users/rmoore/code/tech-utils && go test ./handlers/... -v -run Units
```
Expected: PASS

**Step 5: Create units template**

Create `templates/units.html`:
```html
{{define "content"}}
<h1>Unit Converter</h1>
<p>Convert between different units of measurement.</p>

<form hx-post="/units/convert" hx-target="#result" hx-swap="innerHTML">
    <label>
        Value
        <input type="number" name="value" step="any" required placeholder="Enter a number">
    </label>
    <label>
        Category
        <select name="category" id="category" onchange="updateUnits()">
            {{range .Categories}}
            <option value="{{.Name}}">{{.Name}}</option>
            {{end}}
        </select>
    </label>
    <div class="grid">
        <label>
            From
            <select name="from" id="from-unit">
                {{range (index .Categories 0).Units}}
                <option value="{{.Code}}">{{.Name}} ({{.Code}})</option>
                {{end}}
            </select>
        </label>
        <label>
            To
            <select name="to" id="to-unit">
                {{range (index .Categories 0).Units}}
                <option value="{{.Code}}">{{.Name}} ({{.Code}})</option>
                {{end}}
            </select>
        </label>
    </div>
    <button type="submit">Convert</button>
</form>

<div id="result" class="result-container"></div>

<script>
var unitData = {
    {{range .Categories}}
    "{{.Name}}": [
        {{range .Units}}
        {code: "{{.Code}}", name: "{{.Name}}"},
        {{end}}
    ],
    {{end}}
};

function updateUnits() {
    var category = document.getElementById('category').value;
    var units = unitData[category] || [];
    var fromSelect = document.getElementById('from-unit');
    var toSelect = document.getElementById('to-unit');

    var options = units.map(function(u) {
        return '<option value="' + u.code + '">' + u.name + ' (' + u.code + ')</option>';
    }).join('');

    fromSelect.innerHTML = options;
    toSelect.innerHTML = options;
}
</script>
{{end}}
```

**Step 6: Register routes in main.go**

Add to `main.go`:
```go
unitsHandler := handlers.NewUnitsHandler()
e.GET("/units", unitsHandler.Index)
e.POST("/units/convert", unitsHandler.Convert)
```

**Step 7: Commit**

```bash
git add -A && git commit -m "feat: add unit converter tool"
```

---

## Task 9: Finalize main.go

**Files:**
- Modify: `main.go`

**Step 1: Ensure main.go has all routes**

The final `main.go` should look like:
```go
package main

import (
	"embed"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"os"

	"tech-utils/config"
	"tech-utils/handlers"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

//go:embed templates/*.html
var templatesFS embed.FS

//go:embed static/*
var staticFS embed.FS

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func main() {
	cfg := config.Load()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: cfg.LogLevel,
	}))
	slog.SetDefault(logger)

	e := echo.New()
	e.HideBanner = true

	tmpl := template.Must(template.ParseFS(templatesFS, "templates/*.html"))
	e.Renderer = &Template{templates: tmpl}

	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:   true,
		LogURI:      true,
		LogError:    true,
		HandleError: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			if v.Error == nil {
				logger.Info("request",
					slog.String("uri", v.URI),
					slog.Int("status", v.Status),
				)
			} else {
				logger.Error("request error",
					slog.String("uri", v.URI),
					slog.Int("status", v.Status),
					slog.String("error", v.Error.Error()),
				)
			}
			return nil
		},
	}))
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: cfg.AllowedOrigins,
	}))

	e.GET("/static/*", echo.WrapHandler(http.FileServer(http.FS(staticFS))))

	e.GET("/health", func(c echo.Context) error {
		return c.String(200, "OK")
	})

	homeHandler := handlers.NewHomeHandler()
	e.GET("/", homeHandler.Index)

	uuidHandler := handlers.NewUUIDHandler()
	e.GET("/uuid", uuidHandler.Index)
	e.POST("/uuid/generate", uuidHandler.Generate)

	jsonHandler := handlers.NewJSONHandler()
	e.GET("/json", jsonHandler.Index)
	e.POST("/json/format", jsonHandler.Format)

	convertHandler := handlers.NewConvertHandler()
	e.GET("/convert", convertHandler.Index)
	e.POST("/convert/transform", convertHandler.Transform)

	timezoneHandler := handlers.NewTimezoneHandler()
	e.GET("/timezone", timezoneHandler.Index)
	e.POST("/timezone/convert", timezoneHandler.Convert)

	unitsHandler := handlers.NewUnitsHandler()
	e.GET("/units", unitsHandler.Index)
	e.POST("/units/convert", unitsHandler.Convert)

	logger.Info("starting server", slog.String("port", cfg.Port))
	e.Logger.Fatal(e.Start(":" + cfg.Port))
}
```

**Step 2: Run all tests**

Run:
```bash
cd /Users/rmoore/code/tech-utils && go test ./... -v
```
Expected: All PASS

**Step 3: Verify build**

Run:
```bash
cd /Users/rmoore/code/tech-utils && go build ./...
```
Expected: No errors

**Step 4: Commit if changes were made**

```bash
git add -A && git commit -m "chore: finalize main.go with all routes" || echo "Nothing to commit"
```

---

## Task 10: Dockerfile

**Files:**
- Create: `Dockerfile`
- Create: `.dockerignore`

**Step 1: Create .dockerignore**

Create `.dockerignore`:
```
.git
.gitignore
*.md
docs/
.claude/
```

**Step 2: Create Dockerfile**

Create `Dockerfile`:
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

**Step 3: Verify Docker build**

Run:
```bash
cd /Users/rmoore/code/tech-utils && docker build -t tech-utils:test .
```
Expected: Build succeeds

**Step 4: Test Docker container**

Run:
```bash
docker run -d --name tech-utils-test -p 8081:8080 tech-utils:test && sleep 3 && curl -s http://localhost:8081/health && docker stop tech-utils-test && docker rm tech-utils-test
```
Expected: `OK`

**Step 5: Commit**

```bash
git add -A && git commit -m "feat: add Dockerfile with Chainguard images"
```

---

## Task 11: Final Verification

**Step 1: Run full test suite**

Run:
```bash
cd /Users/rmoore/code/tech-utils && go test ./... -v
```
Expected: All PASS

**Step 2: Run vet**

Run:
```bash
cd /Users/rmoore/code/tech-utils && go vet ./...
```
Expected: No issues

**Step 3: Format code**

Run:
```bash
cd /Users/rmoore/code/tech-utils && go fmt ./...
```

**Step 4: Manual smoke test**

Run:
```bash
cd /Users/rmoore/code/tech-utils && go run . &
sleep 2
echo "Testing endpoints..."
curl -s http://localhost:8080/ | grep -q "Developer Tools" && echo "Home OK"
curl -s http://localhost:8080/uuid | grep -q "UUID Generator" && echo "UUID page OK"
curl -s http://localhost:8080/json | grep -q "JSON Formatter" && echo "JSON page OK"
curl -s http://localhost:8080/convert | grep -q "Format Converter" && echo "Convert page OK"
curl -s http://localhost:8080/timezone | grep -q "Timezone Converter" && echo "Timezone page OK"
curl -s http://localhost:8080/units | grep -q "Unit Converter" && echo "Units page OK"
pkill -f "tech-utils"
echo "All smoke tests passed"
```

**Step 5: Final commit if needed**

```bash
git status && git add -A && git commit -m "chore: format code" || echo "Nothing to commit"
```
