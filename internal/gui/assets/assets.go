// Package assets embeds images used by native GUI windows.
package assets

import _ "embed"

// AboutAvatar is the portrait shown beside the developer in the About window.
//
//go:embed miyagi_shiori.jpg
var AboutAvatar []byte
