//go:build embeddedweb

package webassets

import (
	"io/fs"
	"strings"
	"testing"
)

func TestEmbeddedFrontendFiles(t *testing.T) {
	content := FS()
	if content == nil {
		t.Fatal("embedded frontend is missing")
	}
	index, err := fs.ReadFile(content, "index.html")
	if err != nil {
		t.Fatal(err)
	}
	for _, element := range []string{`<div id="app">`, "/assets/"} {
		if !strings.Contains(string(index), element) {
			t.Errorf("embedded frontend index is missing %q", element)
		}
	}
	entries, err := fs.ReadDir(content, "assets")
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) == 0 {
		t.Fatal("embedded frontend assets are empty")
	}
}
