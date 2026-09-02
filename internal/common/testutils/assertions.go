//go:build unit || integration

package testutils

import (
	"log"
	"reflect"
	"testing"

	"github.com/ohler55/ojg/jp"
)

func AssertEmpty[E any](t *testing.T, result []E) {
	t.Helper()
	if len(result) != 0 {
		t.Errorf("expected empty result, got %d elements: %+v", len(result), result)
	}
}

func AssertContainsInAnyOrder[E any](t *testing.T, result, want []E) {
	t.Helper()

	if len(result) != len(want) {
		t.Fatalf("expected %d elements, got %d", len(want), len(result))
	}

	for _, wanted := range want {
		found := false
		for _, element := range result {
			if reflect.DeepEqual(element, wanted) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected element not found in result:\n  want: %+v\n  got:  %+v", wanted, result)
		}
	}
}

func AssertExpr(t *testing.T, e jp.Expr, wanted string) {
	t.Helper()
	if e.String() != wanted {
		t.Errorf("expression mismatch:\n  got:  %s\n  want: %s", e, wanted)
	}
}

func AssertResolvesToNode(t *testing.T, doc any, e jp.Expr, wanted any) {
	t.Helper()
	if found := e.First(doc); !reflect.DeepEqual(found, wanted) {
		t.Errorf("expression resolves to wrong node:\n  got:  %+v\n  want: %+v", found, wanted)
	}
}

func AssertNoError[T any](value T, err error) T {
	if rerr := reflect.ValueOf(err); rerr.IsValid() && !rerr.IsZero() {
		log.Panicf("unexpected error: %v", err)
	}

	return value
}

func AssertDeepEquals(t *testing.T, expected any, found any) {
	if !reflect.DeepEqual(found, expected) {
		t.Errorf("result does not match expected output\n  got:  %+v\n  want: %+v", found, expected)
	}
}

func AssertPanics(t *testing.T, message string) {
	t.Helper()

	if err := recover(); err == nil {
		t.Fatal(message)
	}
}

func AssertDoesNotPanic(t *testing.T, message string) {
	t.Helper()

	if err := recover(); err != nil {
		t.Fatalf(message, err)
	}
}
