package ui

import (
	"embed"
	"io/fs"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

//go:embed templates/*.html
var testTemplates embed.FS

//go:embed static/*
var testStatic embed.FS

func TestNewUIHandler(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Prepare template FS (files at root of the sub FS)
	tmplFS, err := fs.Sub(testTemplates, "templates")
	assert.NoError(t, err)

	// Prepare static FS
	staticFS, err := fs.Sub(testStatic, "static")
	assert.NoError(t, err)

	h, err := NewUIHandler(tmplFS, staticFS, logger, nil)
	assert.NoError(t, err)
	assert.NotNil(t, h)
	assert.NotNil(t, h.templates)

	// Verify some templates exist
	assert.NotNil(t, h.templates["dashboard.html"])
	assert.NotNil(t, h.templates["searches.html"])
}

func TestParseTemplatesError(t *testing.T) {
	// Test with an empty FS
	emptyFS := embed.FS{}
	_, err := parseTemplates(emptyFS)
	assert.Error(t, err)
}
