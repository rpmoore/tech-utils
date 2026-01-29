package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
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
			err := h.Convert(c)
			require.NoError(t, err, "handler should not return error")

			require.Equal(t, tt.wantStatus, rec.Code, "status code mismatch")

			body := rec.Body.String()
			for _, want := range tt.wantInBody {
				require.Contains(t, body, want, "body should contain expected string")
			}
		})
	}
}
