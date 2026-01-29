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
	return c.Render(http.StatusOK, "uuid.html", map[string]any{
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
	for range count {
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
