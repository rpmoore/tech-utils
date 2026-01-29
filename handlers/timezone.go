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
	return c.Render(http.StatusOK, "timezone.html", map[string]any{
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
