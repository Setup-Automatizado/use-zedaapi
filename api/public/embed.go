package public

import "embed"

//go:embed *.ico *.png *.json *.xml
var Assets embed.FS
