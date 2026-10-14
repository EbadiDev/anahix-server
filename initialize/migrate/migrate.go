package migrate

import (
	"embed"
)

//go:embed database/*.sql
var sqlFiles embed.FS
