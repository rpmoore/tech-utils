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

func (h *HomeHandler) GetPageData() map[string]any {
	tools := []Tool{
		{Name: "UUID Generator", Description: "Generate UUIDs (v4 and v7)", Path: "/uuid"},
		{Name: "JSON Formatter", Description: "Pretty-print and syntax highlight JSON", Path: "/json"},
		{Name: "Format Converter", Description: "Convert between JSON, YAML, and TOML", Path: "/convert"},
		{Name: "Timezone Converter", Description: "Convert times between timezones", Path: "/timezone"},
		{Name: "Unit Converter", Description: "Convert between units of measurement", Path: "/units"},
	}

	return map[string]any{
		"Title": "Home",
		"Tools": tools,
	}
}

func (h *HomeHandler) Index(c echo.Context) error {
	return c.Render(http.StatusOK, "home.html", h.GetPageData())
}
