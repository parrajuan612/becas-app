// Package assets embebe los recursos estáticos del backend.
package assets

import _ "embed"

//go:embed logo.png
var Logo []byte
