//go:build embeddedweb

package webassets

import (
	"embed"
	"io/fs"
)

//go:embed dist
var assets embed.FS

// FS returns the frontend bundled into a production executable.
func FS() fs.FS {
	content, err := fs.Sub(assets, "dist")
	if err != nil {
		return nil
	}
	return content
}
