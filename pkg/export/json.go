package export

import (
	"encoding/json"
	"os"

	"github.com/dshills/dungo/pkg/dungeon"
)

// ExportJSON serializes the complete artifact to JSON with indentation.
// Returns formatted JSON with 2-space indentation for readability.
func ExportJSON(artifact *dungeon.Artifact) ([]byte, error) {
	return json.MarshalIndent(artifact, "", "  ")
}

// ExportJSONCompact serializes the artifact to JSON without indentation.
// Returns compact JSON suitable for storage or transmission.
func ExportJSONCompact(artifact *dungeon.Artifact) ([]byte, error) {
	return json.Marshal(artifact)
}

// SaveJSONToFile exports the artifact to a JSON file with indentation.
// The file is created with 0644 permissions (readable by all, writable by owner).
func SaveJSONToFile(artifact *dungeon.Artifact, filepath string) error {
	data, err := ExportJSON(artifact)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath, data, 0644)
}

// SaveJSONCompactToFile exports the artifact to a compact JSON file.
// The file is created with 0644 permissions (readable by all, writable by owner).
func SaveJSONCompactToFile(artifact *dungeon.Artifact, filepath string) error {
	data, err := ExportJSONCompact(artifact)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath, data, 0644)
}
