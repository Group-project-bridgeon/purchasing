package migrations

import "embed"

// FS embeds all .up.sql migration scripts into the binary.
//
//go:embed *.up.sql
var FS embed.FS
