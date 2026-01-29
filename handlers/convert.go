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
	return c.Render(http.StatusOK, "convert.html", map[string]any{
		"Title": "Format Converter",
	})
}

func (h *ConvertHandler) Transform(c echo.Context) error {
	input := c.FormValue("input")
	fromFormat := c.FormValue("from")
	toFormat := c.FormValue("to")

	var data any
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
