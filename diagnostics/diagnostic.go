package diagnostics

import (
	"fmt"
	"os"
)

// Severity levels for diagnostics
type Severity int

const (
	SeverityError Severity = iota
	SeverityWarning
	SeverityInfo
)

// Diagnostic represents a compiler diagnostic message
type Diagnostic struct {
	Severity Severity
	Message  string
	Line     int
	Column   int
	File     string
}

// DiagnosticEngine collects and reports diagnostics
type DiagnosticEngine struct {
	diagnostics []Diagnostic
	errorCount  int
	warnCount   int
}

// NewDiagnosticEngine creates a new diagnostic engine
func NewDiagnosticEngine() *DiagnosticEngine {
	return &DiagnosticEngine{
		diagnostics: make([]Diagnostic, 0),
	}
}

// Error reports an error
func (d *DiagnosticEngine) Error(message string) {
	d.diagnostics = append(d.diagnostics, Diagnostic{
		Severity: SeverityError,
		Message:  message,
	})
	d.errorCount++
}

// ErrorAt reports an error at a specific location
func (d *DiagnosticEngine) ErrorAt(file string, line, column int, message string) {
	d.diagnostics = append(d.diagnostics, Diagnostic{
		Severity: SeverityError,
		Message:  message,
		File:     file,
		Line:     line,
		Column:   column,
	})
	d.errorCount++
}

// Warning reports a warning
func (d *DiagnosticEngine) Warning(message string) {
	d.diagnostics = append(d.diagnostics, Diagnostic{
		Severity: SeverityWarning,
		Message:  message,
	})
	d.warnCount++
}

// WarningAt reports a warning at a specific location
func (d *DiagnosticEngine) WarningAt(file string, line, column int, message string) {
	d.diagnostics = append(d.diagnostics, Diagnostic{
		Severity: SeverityWarning,
		Message:  message,
		File:     file,
		Line:     line,
		Column:   column,
	})
	d.warnCount++
}

// HasErrors returns true if any errors were reported
func (d *DiagnosticEngine) HasErrors() bool {
	return d.errorCount > 0
}

// ErrorCount returns the number of errors
func (d *DiagnosticEngine) ErrorCount() int {
	return d.errorCount
}

// WarningCount returns the number of warnings
func (d *DiagnosticEngine) WarningCount() int {
	return d.warnCount
}

// Print outputs all diagnostics
func (d *DiagnosticEngine) Print() {
	for _, diag := range d.diagnostics {
		var prefix string
		switch diag.Severity {
		case SeverityError:
			prefix = "ERROR"
		case SeverityWarning:
			prefix = "WARNING"
		case SeverityInfo:
			prefix = "INFO"
		}
		
		if diag.File != "" {
			fmt.Fprintf(os.Stderr, "%s:%d:%d: %s: %s\n", 
				diag.File, diag.Line, diag.Column, prefix, diag.Message)
		} else {
			fmt.Fprintf(os.Stderr, "%s: %s\n", prefix, diag.Message)
		}
	}
}