package errors

import (
	"fmt"
	"math"
	"reflect"
	"unsafe"

	goerrors "github.com/go-errors/errors"
	"github.com/open-resource-discovery/overlay-golang/internal/common/utils"
)

type Severity int

const (
	Severity_Warning Severity = iota
	Severity_Error
	Severity_Fatal
)

type OverlayError struct {
	severity Severity
	errors   []error
}

func Create(severity Severity, message string, args ...any) *OverlayError {
	return &OverlayError{
		severity: severity,
		errors:   []error{goerrors.Wrap(fmt.Errorf(message, args...), 1)},
	}
}

func Wrap(err any, severity ...Severity) *OverlayError {
	if rerr := reflect.ValueOf(err); !rerr.IsValid() || rerr.IsZero() {
		return nil
	}

	if oerr, ok := err.(*OverlayError); ok {
		return &OverlayError{
			errors:   append([]error{}, oerr.errors...),
			severity: append(severity, oerr.Severity())[0],
		}
	}

	return &OverlayError{
		errors:   []error{fmt.Errorf("%v", err)},
		severity: append(severity, Severity_Error)[0],
	}
}

func WrapPrefix(err any, prefix string, args ...any) *OverlayError {
	if wrapped := Wrap(err); wrapped == nil {
		return nil
	} else {
		return &OverlayError{
			severity: wrapped.Severity(),
			errors:   []error{goerrors.WrapPrefix(wrapped, fmt.Sprintf(prefix, args...), 1)},
		}
	}
}

func (self *OverlayError) Error() string {
	if len(self.errors) == 1 {
		return self.errors[0].Error()
	}

	b := []byte(self.errors[0].Error())

	for _, err := range self.errors[1:] {
		b = append(b, '\n')
		b = append(b, err.Error()...)
	}

	return unsafe.String(&b[0], len(b))
}

func (self *OverlayError) Unwrap() []error {
	return append([]error{}, self.errors...)
}

func (self *OverlayError) Severity() Severity {
	return self.severity
}

func Append(destination *OverlayError, source any) *OverlayError {
	wrapped := Wrap(source)

	if destination != nil && wrapped != nil {
		destination.errors = append(destination.errors, wrapped)
		destination.severity = Severity(math.Max(float64(destination.severity), float64(wrapped.Severity())))
	}

	return utils.Ternary(destination != nil, destination, wrapped)
}
