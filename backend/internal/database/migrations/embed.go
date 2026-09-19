package migrations

import "embed"

// FS embeds all canonical SQL migration files into the compiled Go binary.
// This ensures that migration discovery works reliably in containers and binaries
// without requiring relative filesystem paths.
//
//go:embed *.sql
var FS embed.FS
