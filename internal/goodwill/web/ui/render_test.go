package ui

import (
	"bytes"
	"io/fs"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestTemplateRenderingWithFlashMessages(t *testing.T) {
	logger := logrus.New()

	tmplFS, err := fs.Sub(testTemplates, "templates")
	assert.NoError(t, err)

	staticFS, err := fs.Sub(testStatic, "static")
	assert.NoError(t, err)

	h, err := NewUIHandler(tmplFS, staticFS, logger, nil)
	assert.NoError(t, err)

	testCases := []struct {
		name     string
		template string
		data     interface{}
	}{
		{
			name:     "Dashboard",
			template: "dashboard.html",
			data: DashboardData{
				Title:         "Test",
				FlashMessages: []FlashMessage{{Type: "success", Message: "Hello"}},
			},
		},
		{
			name:     "Searches",
			template: "searches.html",
			data: SearchesData{
				Title:         "Test",
				FlashMessages: []FlashMessage{{Type: "danger", Message: "Error"}},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			tmpl, ok := h.templates[tc.template]
			assert.True(t, ok)
			err := tmpl.Execute(&buf, tc.data)
			assert.NoError(t, err)
			assert.Contains(t, buf.String(), "Test")
			if tc.name == "Dashboard" {
				assert.Contains(t, buf.String(), "success")
				assert.Contains(t, buf.String(), "Hello")
			} else if tc.name == "Searches" {
				assert.Contains(t, buf.String(), "danger")
				assert.Contains(t, buf.String(), "Error")
			}
		})
	}
}
