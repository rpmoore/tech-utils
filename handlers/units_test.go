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
