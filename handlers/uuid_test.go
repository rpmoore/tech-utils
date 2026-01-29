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
