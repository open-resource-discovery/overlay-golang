//go:build unit

package errors

import (
	"errors"
	"testing"
)

// ---- Create -----------------------------------------------------------------

func TestCreate_ReturnsNonNil(t *testing.T) {
	err := Create(Severity_Error, "something went wrong")
	if err == nil {
		t.Fatal("Create returned nil")
	}
}

func TestCreate_ErrorMessageFormatted(t *testing.T) {
	err := Create(Severity_Error, "value is %d", 42)
	if err.Error() != "value is 42" {
		t.Errorf("Error() = %q, want %q", err.Error(), "value is 42")
	}
}

func TestCreate_SeverityStored(t *testing.T) {
	cases := []struct {
		severity Severity
	}{
		{Severity_Warning},
		{Severity_Error},
		{Severity_Fatal},
	}
	for _, tc := range cases {
		err := Create(tc.severity, "msg")
		if err.Severity() != tc.severity {
			t.Errorf("Severity() = %v, want %v", err.Severity(), tc.severity)
		}
	}
}

func TestCreate_SingleErrorInUnwrap(t *testing.T) {
	err := Create(Severity_Error, "msg")
	if len(err.Unwrap()) != 1 {
		t.Fatalf("Unwrap() len = %d, want 1", len(err.Unwrap()))
	}
}

// ---- Wrap -------------------------------------------------------------------

func TestWrap_NilInput_ReturnsNil(t *testing.T) {
	if got := Wrap(nil); got != nil {
		t.Errorf("Wrap(nil) = %v, want nil", got)
	}
}

func TestWrap_ZeroValueError_ReturnsNil(t *testing.T) {
	var err error
	if got := Wrap(err); got != nil {
		t.Errorf("Wrap(zero error) = %v, want nil", got)
	}
}

func TestWrap_PlainError_DefaultsSeverityToError(t *testing.T) {
	wrapped := Wrap(errors.New("plain error"))
	if wrapped == nil {
		t.Fatal("Wrap returned nil for non-nil error")
	}
	if wrapped.Severity() != Severity_Error {
		t.Errorf("Severity() = %v, want Severity_Error", wrapped.Severity())
	}
}

func TestWrap_PlainError_MessagePreserved(t *testing.T) {
	wrapped := Wrap(errors.New("original message"))
	if wrapped.Error() != "original message" {
		t.Errorf("Error() = %q, want %q", wrapped.Error(), "original message")
	}
}

func TestWrap_PlainError_SeverityOverride(t *testing.T) {
	wrapped := Wrap(errors.New("msg"), Severity_Warning)
	if wrapped.Severity() != Severity_Warning {
		t.Errorf("Severity() = %v, want Severity_Warning", wrapped.Severity())
	}
}

func TestWrap_OverlayError_CopiesErrors(t *testing.T) {
	original := Create(Severity_Warning, "original")
	wrapped := Wrap(original)
	if wrapped == nil {
		t.Fatal("Wrap(OverlayError) returned nil")
	}
	if len(wrapped.Unwrap()) != len(original.Unwrap()) {
		t.Errorf("Unwrap() len = %d, want %d", len(wrapped.Unwrap()), len(original.Unwrap()))
	}
}

func TestWrap_OverlayError_PreservesSeverityWhenNoOverride(t *testing.T) {
	original := Create(Severity_Fatal, "fatal")
	wrapped := Wrap(original)
	if wrapped.Severity() != Severity_Fatal {
		t.Errorf("Severity() = %v, want Severity_Fatal", wrapped.Severity())
	}
}

func TestWrap_OverlayError_SeverityOverrideApplied(t *testing.T) {
	original := Create(Severity_Fatal, "fatal")
	wrapped := Wrap(original, Severity_Warning)
	if wrapped.Severity() != Severity_Warning {
		t.Errorf("Severity() = %v, want Severity_Warning", wrapped.Severity())
	}
}

func TestWrap_OverlayError_DoesNotMutateOriginal(t *testing.T) {
	original := Create(Severity_Error, "original")
	wrapped := Wrap(original, Severity_Warning)
	_ = wrapped
	if original.Severity() != Severity_Error {
		t.Error("Wrap mutated the original OverlayError's severity")
	}
}

// ---- WrapPrefix -------------------------------------------------------------

func TestWrapPrefix_NilInput_ReturnsNil(t *testing.T) {
	if got := WrapPrefix(nil, "prefix"); got != nil {
		t.Errorf("WrapPrefix(nil) = %v, want nil", got)
	}
}

func TestWrapPrefix_ZeroError_ReturnsNil(t *testing.T) {
	var err error
	if got := WrapPrefix(err, "prefix"); got != nil {
		t.Errorf("WrapPrefix(zero error) = %v, want nil", got)
	}
}

func TestWrapPrefix_PlainError_PrefixInMessage(t *testing.T) {
	wrapped := WrapPrefix(errors.New("detail"), "context")
	if wrapped == nil {
		t.Fatal("WrapPrefix returned nil for non-nil error")
	}
	msg := wrapped.Error()
	if msg == "" {
		t.Error("WrapPrefix produced empty message")
	}
}

func TestWrapPrefix_FormattedPrefix(t *testing.T) {
	wrapped := WrapPrefix(errors.New("detail"), "step %d failed", 3)
	if wrapped == nil {
		t.Fatal("WrapPrefix returned nil")
	}
}

func TestWrapPrefix_OverlayError_PreservesSeverity(t *testing.T) {
	original := Create(Severity_Warning, "base error")
	wrapped := WrapPrefix(original, "outer context")
	if wrapped.Severity() != Severity_Warning {
		t.Errorf("Severity() = %v, want Severity_Warning", wrapped.Severity())
	}
}

// ---- Error ------------------------------------------------------------------

func TestError_SingleError_ReturnsMessage(t *testing.T) {
	err := Create(Severity_Error, "single error message")
	if err.Error() != "single error message" {
		t.Errorf("Error() = %q, want %q", err.Error(), "single error message")
	}
}

func TestError_MultipleErrors_JoinedWithNewline(t *testing.T) {
	base := Create(Severity_Error, "first")
	Append(base, errors.New("second"))

	got := base.Error()
	want := "first\nsecond"
	if got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

// ---- Unwrap -----------------------------------------------------------------

func TestUnwrap_ReturnsUnderlyingErrors(t *testing.T) {
	err := Create(Severity_Error, "msg")
	errs := err.Unwrap()
	if len(errs) == 0 {
		t.Fatal("Unwrap() returned empty slice")
	}
}

func TestUnwrap_MultipleErrors_AllPresent(t *testing.T) {
	base := Create(Severity_Error, "first")
	Append(base, errors.New("second"))
	Append(base, errors.New("third"))

	if len(base.Unwrap()) != 3 {
		t.Errorf("Unwrap() len = %d, want 3", len(base.Unwrap()))
	}
}

// ---- Append -----------------------------------------------------------------

func TestAppend_NilReceiver_ReturnsWrappedArgument(t *testing.T) {
	var base *OverlayError
	result := Append(base, errors.New("first error"))
	if result == nil {
		t.Fatal("Append on nil receiver returned nil")
	}
	if result.Error() != "first error" {
		t.Errorf("Error() = %q, want %q", result.Error(), "first error")
	}
}

func TestAppend_NilReceiver_NilArgument_ReturnsNil(t *testing.T) {
	var base *OverlayError
	result := Append(base, nil)
	if result != nil {
		t.Errorf("Append(nil) on nil receiver = %v, want nil", result)
	}
}

func TestAppend_NonNilReceiver_AddsError(t *testing.T) {
	base := Create(Severity_Error, "first")
	result := Append(base, errors.New("second"))

	if result != base {
		t.Error("Append on non-nil receiver should return same receiver")
	}
	if len(base.Unwrap()) != 2 {
		t.Errorf("Unwrap() len = %d, want 2", len(base.Unwrap()))
	}
}

func TestAppend_NilArgument_NoChange(t *testing.T) {
	base := Create(Severity_Error, "only")
	result := Append(base, nil)

	if result != base {
		t.Error("Append(nil) should return the receiver unchanged")
	}
	if len(base.Unwrap()) != 1 {
		t.Errorf("Unwrap() len = %d, want 1 after Append(nil)", len(base.Unwrap()))
	}
}

func TestAppend_SeverityEscalates_WhenAppendedIsHigher(t *testing.T) {
	base := Create(Severity_Warning, "warning")
	Append(base, Create(Severity_Fatal, "fatal"))

	if base.Severity() != Severity_Fatal {
		t.Errorf("Severity() = %v, want Severity_Fatal after appending fatal", base.Severity())
	}
}

func TestAppend_SeverityUnchanged_WhenAppendedIsLower(t *testing.T) {
	base := Append(Create(Severity_Fatal, "fatal"), Create(Severity_Warning, "warning"))

	if base.Severity() != Severity_Fatal {
		t.Errorf("Severity() = %v, want Severity_Fatal (unchanged)", base.Severity())
	}
}

func TestAppend_SeverityUnchanged_WhenAppendedIsEqual(t *testing.T) {
	base := Append(Create(Severity_Error, "first"), Create(Severity_Error, "second"))

	if base.Severity() != Severity_Error {
		t.Errorf("Severity() = %v, want Severity_Error", base.Severity())
	}
}

func TestAppend_AppendOverlayError_WrapsAsEntry(t *testing.T) {
	base := Append(Create(Severity_Error, "base"), Create(Severity_Warning, "extra"))

	if len(base.Unwrap()) != 2 {
		t.Errorf("Unwrap() len = %d, want 2", len(base.Unwrap()))
	}
}

// ---- Severity ordering ------------------------------------------------------

func TestSeverity_Ordering(t *testing.T) {
	if Severity_Warning >= Severity_Error {
		t.Error("Severity_Warning must be less than Severity_Error")
	}
	if Severity_Error >= Severity_Fatal {
		t.Error("Severity_Error must be less than Severity_Fatal")
	}
}

// ---- errors.Is / errors.As integration --------------------------------------

func TestErrorsAs_WrappedPlainError_FindsIt(t *testing.T) {
	inner := errors.New("inner")
	wrapped := Wrap(inner)

	// errors.As walks the Unwrap() chain; one of the leaves should be *OverlayError.
	var target *OverlayError
	if !errors.As(wrapped, &target) {
		t.Error("errors.As should find *OverlayError in the chain")
	}
}

func TestErrorsIs_MultipleErrors_AppendWrapsAsSeparateEntry(t *testing.T) {
	// Append wraps the argument into an *OverlayError, so the original sentinel
	// is not reachable via errors.Is identity — this test documents that boundary.
	sentinel := errors.New("sentinel")
	base := Append(Create(Severity_Error, "first"), sentinel)

	// The wrapped entry is an *OverlayError, not the sentinel itself.
	if errors.Is(base, sentinel) {
		t.Error("Append wraps the error — sentinel identity should not be reachable via errors.Is")
	}
	// The message is preserved in the chain, however.
	if base.Error() != "first\nsentinel" {
		t.Errorf("Error() = %q, want \"first\\nsentinel\"", base.Error())
	}
}
