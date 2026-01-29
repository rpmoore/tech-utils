package handlers

import (
	"testing"
)

func TestHomeHandler_GetPageData(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		wantTitle     string
		wantToolCount int
		wantTools     []string
	}{
		"returns correct page data": {
			wantTitle:     "Home",
			wantToolCount: 5,
			wantTools:     []string{"UUID Generator", "JSON Formatter", "Format Converter", "Timezone Converter", "Unit Converter"},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			h := &HomeHandler{}
			data := h.GetPageData()

			if data["Title"] != tt.wantTitle {
				t.Errorf("Title = %v, want %v", data["Title"], tt.wantTitle)
			}

			tools, ok := data["Tools"].([]Tool)
			if !ok {
				t.Fatal("Tools is not []Tool")
			}

			if len(tools) != tt.wantToolCount {
				t.Errorf("len(Tools) = %d, want %d", len(tools), tt.wantToolCount)
			}

			for i, tool := range tools {
				if tool.Name != tt.wantTools[i] {
					t.Errorf("Tools[%d].Name = %s, want %s", i, tool.Name, tt.wantTools[i])
				}
			}
		})
	}
}
