package validation

import (
	"encoding/json"
	"os"

	"github.com/dshills/dungo/pkg/dungeon"
)

// ExportReportJSON serializes a ValidationReport to JSON with indentation.
// Returns formatted JSON with 2-space indentation for readability.
func ExportReportJSON(report *dungeon.ValidationReport) ([]byte, error) {
	return json.MarshalIndent(report, "", "  ")
}

// ExportReportJSONCompact serializes a ValidationReport to JSON without indentation.
// Returns compact JSON suitable for storage or transmission.
func ExportReportJSONCompact(report *dungeon.ValidationReport) ([]byte, error) {
	return json.Marshal(report)
}

// SaveReportToFile exports a ValidationReport to a JSON file with indentation.
// The file is created with 0644 permissions (readable by all, writable by owner).
func SaveReportToFile(report *dungeon.ValidationReport, filepath string) error {
	data, err := ExportReportJSON(report)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath, data, 0644)
}

// SaveReportCompactToFile exports a ValidationReport to a compact JSON file.
// The file is created with 0644 permissions (readable by all, writable by owner).
func SaveReportCompactToFile(report *dungeon.ValidationReport, filepath string) error {
	data, err := ExportReportJSONCompact(report)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath, data, 0644)
}

// LoadReportFromFile loads a ValidationReport from a JSON file.
func LoadReportFromFile(filepath string) (*dungeon.ValidationReport, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	var report dungeon.ValidationReport
	if err := json.Unmarshal(data, &report); err != nil {
		return nil, err
	}

	return &report, nil
}
