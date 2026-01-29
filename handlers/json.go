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
	return c.Render(http.StatusOK, "json.html", map[string]any{
		"Title": "JSON Formatter",
	})
}

func (h *JSONHandler) Format(c echo.Context) error {
	input := c.FormValue("input")
	indentType := c.FormValue("indent")

	var parsed any
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
	// Only escape characters that could cause XSS, but leave quotes intact for JSON display
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")

	// Highlight strings first, using a placeholder to protect them
	stringRe := regexp.MustCompile(`"((?:[^"\\]|\\.)*)"`)
	var stringMatches []string
	s = stringRe.ReplaceAllStringFunc(s, func(match string) string {
		idx := len(stringMatches)
		stringMatches = append(stringMatches, match)
		return "___STRING_" + string(rune('A'+idx)) + "___"
	})

	// Now highlight numbers, booleans, and null (won't match inside strings)
	numberRe := regexp.MustCompile(`\b(-?\d+\.?\d*)\b`)
	s = numberRe.ReplaceAllString(s, `<span class="syntax-number">$1</span>`)

	boolRe := regexp.MustCompile(`\b(true|false)\b`)
	s = boolRe.ReplaceAllString(s, `<span class="syntax-boolean">$1</span>`)

	nullRe := regexp.MustCompile(`\bnull\b`)
	s = nullRe.ReplaceAllString(s, `<span class="syntax-null">null</span>`)

	// Restore strings with highlighting
	for idx, str := range stringMatches {
		placeholder := "___STRING_" + string(rune('A'+idx)) + "___"
		highlighted := `<span class="syntax-string">` + str + `</span>`
		s = strings.ReplaceAll(s, placeholder, highlighted)
	}

	return s
}
