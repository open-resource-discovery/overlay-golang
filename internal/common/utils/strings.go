package utils

import (
	"strings"
)

func Join(separator string, elements ...string) string {
	return strings.Join(
		Filter(elements, func(s string) bool { return len(s) > 0 }),
		separator,
	)
}
