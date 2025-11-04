package validation_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/dshills/dungo/pkg/dungeon"
	"github.com/dshills/dungo/pkg/validation"
)

// createTestReport creates a test validation report.
func createTestReport() *dungeon.ValidationReport {
	return &dungeon.ValidationReport{
		Passed: true,
		HardConstraintResults: []dungeon.ConstraintResult{
			{
				Constraint: &dungeon.Constraint{
					Kind:     "connectivity",
					Severity: "hard",
					Expr:     "all_rooms_connected",
				},
				Satisfied: true,
				Score:     1.0,
				Details:   "All rooms are reachable from start",
			},
			{
				Constraint: &dungeon.Constraint{
					Kind:     "key_before_lock",
					Severity: "hard",
					Expr:     "keys_accessible",
				},
				Satisfied: true,
				Score:     1.0,
				Details:   "All keys are reachable before their locks",
			},
		},
		SoftConstraintResults: []dungeon.ConstraintResult{
			{
				Constraint: &dungeon.Constraint{
					Kind:     "pacing",
					Severity: "soft",
					Expr:     "difficulty_curve",
				},
				Satisfied: true,
				Score:     0.92,
				Details:   "Pacing deviation: 0.08",
			},
		},
		Metrics: &dungeon.Metrics{
			BranchingFactor:   1.8,
			PathLength:        10,
			CycleCount:        2,
			PacingDeviation:   0.08,
			SecretFindability: 0.85,
		},
		Warnings: []string{
			"Room 'treasury' has high loot density",
		},
		Errors: []string{},
	}
}

// TestExportReportJSON tests JSON export of validation reports.
func TestExportReportJSON(t *testing.T) {
	report := createTestReport()

	data, err := validation.ExportReportJSON(report)
	if err != nil {
		t.Fatalf("ExportReportJSON failed: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("ExportReportJSON returned empty data")
	}

	// Verify it's valid JSON
	var result dungeon.ValidationReport
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Exported JSON is invalid: %v", err)
	}

	// Verify structure
	if !result.Passed {
		t.Error("Passed flag not preserved")
	}

	if len(result.HardConstraintResults) != len(report.HardConstraintResults) {
		t.Errorf("Hard constraint count mismatch: got %d, want %d",
			len(result.HardConstraintResults), len(report.HardConstraintResults))
	}

	if len(result.SoftConstraintResults) != len(report.SoftConstraintResults) {
		t.Errorf("Soft constraint count mismatch: got %d, want %d",
			len(result.SoftConstraintResults), len(report.SoftConstraintResults))
	}
}

// TestExportReportJSONCompact tests compact JSON export.
func TestExportReportJSONCompact(t *testing.T) {
	report := createTestReport()

	compact, err := validation.ExportReportJSONCompact(report)
	if err != nil {
		t.Fatalf("ExportReportJSONCompact failed: %v", err)
	}

	pretty, err := validation.ExportReportJSON(report)
	if err != nil {
		t.Fatalf("ExportReportJSON failed: %v", err)
	}

	// Compact should be smaller
	if len(compact) >= len(pretty) {
		t.Errorf("Compact JSON (%d bytes) is not smaller than pretty JSON (%d bytes)",
			len(compact), len(pretty))
	}
}

// TestReportJSONRoundTrip tests export and re-import.
func TestReportJSONRoundTrip(t *testing.T) {
	original := createTestReport()

	// Export
	data, err := validation.ExportReportJSON(original)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Import
	var restored dungeon.ValidationReport
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("Import failed: %v", err)
	}

	// Verify critical fields
	if restored.Passed != original.Passed {
		t.Errorf("Passed mismatch: got %v, want %v", restored.Passed, original.Passed)
	}

	if len(restored.HardConstraintResults) != len(original.HardConstraintResults) {
		t.Errorf("Hard constraints count mismatch: got %d, want %d",
			len(restored.HardConstraintResults), len(original.HardConstraintResults))
	}

	if len(restored.Warnings) != len(original.Warnings) {
		t.Errorf("Warnings count mismatch: got %d, want %d",
			len(restored.Warnings), len(original.Warnings))
	}

	if restored.Metrics != nil && original.Metrics != nil {
		if restored.Metrics.BranchingFactor != original.Metrics.BranchingFactor {
			t.Errorf("BranchingFactor mismatch: got %f, want %f",
				restored.Metrics.BranchingFactor, original.Metrics.BranchingFactor)
		}
	}
}

// TestSaveReportToFile tests saving report to file.
func TestSaveReportToFile(t *testing.T) {
	report := createTestReport()
	tmpDir := t.TempDir()
	filepath := filepath.Join(tmpDir, "report.json")

	if err := validation.SaveReportToFile(report, filepath); err != nil {
		t.Fatalf("SaveReportToFile failed: %v", err)
	}

	// Verify file exists
	info, err := os.Stat(filepath)
	if err != nil {
		t.Fatalf("Output file not found: %v", err)
	}

	if info.Size() == 0 {
		t.Error("Output file is empty")
	}

	// Verify content
	data, err := os.ReadFile(filepath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	var result dungeon.ValidationReport
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Output file contains invalid JSON: %v", err)
	}
}

// TestSaveReportCompactToFile tests saving compact report.
func TestSaveReportCompactToFile(t *testing.T) {
	report := createTestReport()
	tmpDir := t.TempDir()
	compactPath := filepath.Join(tmpDir, "report_compact.json")
	prettyPath := filepath.Join(tmpDir, "report_pretty.json")

	if err := validation.SaveReportCompactToFile(report, compactPath); err != nil {
		t.Fatalf("SaveReportCompactToFile failed: %v", err)
	}

	if err := validation.SaveReportToFile(report, prettyPath); err != nil {
		t.Fatalf("SaveReportToFile failed: %v", err)
	}

	// Compare sizes
	compactInfo, err := os.Stat(compactPath)
	if err != nil {
		t.Fatalf("Compact file not found: %v", err)
	}

	prettyInfo, err := os.Stat(prettyPath)
	if err != nil {
		t.Fatalf("Pretty file not found: %v", err)
	}

	if compactInfo.Size() >= prettyInfo.Size() {
		t.Errorf("Compact file (%d bytes) is not smaller than pretty file (%d bytes)",
			compactInfo.Size(), prettyInfo.Size())
	}
}

// TestLoadReportFromFile tests loading report from file.
func TestLoadReportFromFile(t *testing.T) {
	original := createTestReport()
	tmpDir := t.TempDir()
	filepath := filepath.Join(tmpDir, "report.json")

	// Save
	if err := validation.SaveReportToFile(original, filepath); err != nil {
		t.Fatalf("SaveReportToFile failed: %v", err)
	}

	// Load
	loaded, err := validation.LoadReportFromFile(filepath)
	if err != nil {
		t.Fatalf("LoadReportFromFile failed: %v", err)
	}

	// Verify
	if loaded.Passed != original.Passed {
		t.Errorf("Passed mismatch: got %v, want %v", loaded.Passed, original.Passed)
	}

	if len(loaded.HardConstraintResults) != len(original.HardConstraintResults) {
		t.Errorf("Hard constraints mismatch: got %d, want %d",
			len(loaded.HardConstraintResults), len(original.HardConstraintResults))
	}
}

// TestLoadReportFromFileErrors tests error handling.
func TestLoadReportFromFileErrors(t *testing.T) {
	tests := []struct {
		name     string
		filepath string
		wantErr  bool
	}{
		{
			name:     "nonexistent file",
			filepath: "/nonexistent/file.json",
			wantErr:  true,
		},
		{
			name:     "empty path",
			filepath: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := validation.LoadReportFromFile(tt.filepath)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadReportFromFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestReportMetricsPreservation tests metrics preservation.
func TestReportMetricsPreservation(t *testing.T) {
	report := &dungeon.ValidationReport{
		Passed: true,
		Metrics: &dungeon.Metrics{
			BranchingFactor:   2.5,
			PathLength:        15,
			CycleCount:        3,
			PacingDeviation:   0.123,
			SecretFindability: 0.789,
		},
	}

	data, err := validation.ExportReportJSON(report)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	var restored dungeon.ValidationReport
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("Import failed: %v", err)
	}

	if restored.Metrics == nil {
		t.Fatal("Metrics is nil after restoration")
	}

	// Check each metric
	tests := []struct {
		name string
		got  interface{}
		want interface{}
	}{
		{"BranchingFactor", restored.Metrics.BranchingFactor, report.Metrics.BranchingFactor},
		{"PathLength", restored.Metrics.PathLength, report.Metrics.PathLength},
		{"CycleCount", restored.Metrics.CycleCount, report.Metrics.CycleCount},
		{"PacingDeviation", restored.Metrics.PacingDeviation, report.Metrics.PacingDeviation},
		{"SecretFindability", restored.Metrics.SecretFindability, report.Metrics.SecretFindability},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("%s mismatch: got %v, want %v", tt.name, tt.got, tt.want)
			}
		})
	}
}

// TestReportConstraintPreservation tests constraint preservation.
func TestReportConstraintPreservation(t *testing.T) {
	report := &dungeon.ValidationReport{
		Passed: false,
		HardConstraintResults: []dungeon.ConstraintResult{
			{
				Constraint: &dungeon.Constraint{
					Kind:     "connectivity",
					Severity: "hard",
					Expr:     "all_connected",
				},
				Satisfied: false,
				Score:     0.0,
				Details:   "Room 'x' is unreachable",
			},
		},
		SoftConstraintResults: []dungeon.ConstraintResult{
			{
				Constraint: &dungeon.Constraint{
					Kind:     "pacing",
					Severity: "soft",
					Expr:     "curve_match",
				},
				Satisfied: true,
				Score:     0.85,
				Details:   "Good pacing curve match",
			},
		},
	}

	data, err := validation.ExportReportJSON(report)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	var restored dungeon.ValidationReport
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("Import failed: %v", err)
	}

	// Check hard constraints
	if len(restored.HardConstraintResults) != 1 {
		t.Fatalf("Expected 1 hard constraint, got %d", len(restored.HardConstraintResults))
	}

	hard := restored.HardConstraintResults[0]
	if hard.Satisfied {
		t.Error("Hard constraint should be unsatisfied")
	}
	if hard.Score != 0.0 {
		t.Errorf("Hard constraint score: got %f, want 0.0", hard.Score)
	}

	// Check soft constraints
	if len(restored.SoftConstraintResults) != 1 {
		t.Fatalf("Expected 1 soft constraint, got %d", len(restored.SoftConstraintResults))
	}

	soft := restored.SoftConstraintResults[0]
	if !soft.Satisfied {
		t.Error("Soft constraint should be satisfied")
	}
	if soft.Score != 0.85 {
		t.Errorf("Soft constraint score: got %f, want 0.85", soft.Score)
	}
}

// TestReportWithEmptyFields tests reports with empty fields.
func TestReportWithEmptyFields(t *testing.T) {
	report := &dungeon.ValidationReport{
		Passed:                true,
		HardConstraintResults: []dungeon.ConstraintResult{},
		SoftConstraintResults: []dungeon.ConstraintResult{},
		Warnings:              []string{},
		Errors:                []string{},
		Metrics:               nil,
	}

	data, err := validation.ExportReportJSON(report)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	var restored dungeon.ValidationReport
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("Import failed: %v", err)
	}

	if !restored.Passed {
		t.Error("Passed flag not preserved")
	}
}

// BenchmarkExportReportJSON benchmarks report export.
func BenchmarkExportReportJSON(b *testing.B) {
	report := createTestReport()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := validation.ExportReportJSON(report)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkSaveReportToFile benchmarks saving to file.
func BenchmarkSaveReportToFile(b *testing.B) {
	report := createTestReport()
	tmpDir := b.TempDir()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		filepath := filepath.Join(tmpDir, "report"+string(rune(i))+".json")
		if err := validation.SaveReportToFile(report, filepath); err != nil {
			b.Fatal(err)
		}
	}
}
