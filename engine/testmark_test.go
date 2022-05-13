package datalarkengine

import (
	"path/filepath"
	"testing"
)

func TestSpecs(t *testing.T) {
	matches, _ := filepath.Glob("../docs/using-*.md")
	for _, match := range matches {
		t.Run(filepath.Base(match), func(t *testing.T) {
			testFixture(t, match)
		})
	}
}
