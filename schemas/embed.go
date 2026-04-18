package schemas

import "embed"

// FS contains the committed DevArch V2 JSON schemas.
//
//go:embed *.json
var FS embed.FS

// ReadFile reads an embedded schema file by name.
func ReadFile(name string) ([]byte, error) {
	return FS.ReadFile(name)
}
