package model

type DiagnosticSeverity string

const (
	DiagnosticSeverityWarning DiagnosticSeverity = "warning"
	DiagnosticSeverityError   DiagnosticSeverity = "error"
)

// Diagnostic describes a warning or error produced while applying an overlay.
// OverlayIndex and PatchIndex are zero-based.
type Diagnostic struct {
	Severity     DiagnosticSeverity
	OverlayIndex int
	PatchIndex   int
	Message      string
}

type DiagnosticHandler func(Diagnostic)
