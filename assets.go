// Package assets embeds the build-time artifacts the binary serves at runtime.
// It lives at the module root on purpose: go:embed directives cannot traverse
// upward ("..") into parent directories, so the embed must sit at or above the
// embedded tree. The Vue build output (web/dist) is rooted here, so the embed
// belongs here too — not in internal/handler as an earlier sketch suggested.
//
// NOTE: `web/dist` must exist at build time (run `npm run build` in web/);
// otherwise `go build` fails with "no matching files found".
package assets

import (
	"embed"
	"io/fs"
)

//go:embed all:web/dist
var webDist embed.FS

// WebFS returns the built Vue SPA rooted at its top level, ready to hand to
// http.FileServer.
func WebFS() (fs.FS, error) {
	return fs.Sub(webDist, "web/dist")
}
