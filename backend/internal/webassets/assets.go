//go:build !embeddedweb

package webassets

import "io/fs"

// FS returns nil in development builds; the server can load a local webdist.
func FS() fs.FS { return nil }
