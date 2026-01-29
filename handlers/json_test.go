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
