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
			wantInBody: []string{"name", "test", "42"},
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
			err := h.Transform(c)
			require.NoError(t, err, "handler returned error")

			require.Equal(t, tt.wantStatus, rec.Code, "unexpected status code")

			body := rec.Body.String()
			for _, want := range tt.wantInBody {
				require.Contains(t, body, want, "body missing expected content")
			}
		})
	}
}
