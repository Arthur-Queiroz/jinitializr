package template

import "embed"

// templateFS holds every .tmpl file the renderer can execute. The templates
// live under this package's own directory (internal/template/templates) rather
// than at the module root because go:embed cannot reference parent directories
// ("..") — the embedded tree must sit at or below the embedding file.
//
//go:embed all:templates
var templateFS embed.FS
