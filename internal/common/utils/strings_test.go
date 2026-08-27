//go:build unit

package utils

import "testing"

func TestJoin(t *testing.T) {
	tests := []struct {
		name      string
		separator string
		elements  []string
		want      string
	}{
		{
			name:      "non-empty elements joined with separator",
			separator: "/",
			elements:  []string{"a", "b", "c"},
			want:      "a/b/c",
		},
		{
			name:      "empty strings filtered out",
			separator: "/",
			elements:  []string{"a", "", "c"},
			want:      "a/c",
		},
		{
			name:      "all empty strings returns empty",
			separator: "/",
			elements:  []string{"", "", ""},
			want:      "",
		},
		{
			name:      "no elements returns empty",
			separator: "/",
			elements:  []string{},
			want:      "",
		},
		{
			name:      "single non-empty element returns element",
			separator: "/",
			elements:  []string{"only"},
			want:      "only",
		},
		{
			name:      "single empty element returns empty",
			separator: "/",
			elements:  []string{""},
			want:      "",
		},
		{
			name:      "leading empty element filtered",
			separator: ".",
			elements:  []string{"", "b", "c"},
			want:      "b.c",
		},
		{
			name:      "trailing empty element filtered",
			separator: ".",
			elements:  []string{"a", "b", ""},
			want:      "a.b",
		},
		{
			name:      "multiple consecutive empty elements filtered",
			separator: "-",
			elements:  []string{"x", "", "", "y"},
			want:      "x-y",
		},
		{
			name:      "empty separator concatenates elements",
			separator: "",
			elements:  []string{"foo", "bar", "baz"},
			want:      "foobarbaz",
		},
		{
			name:      "multi-char separator",
			separator: ", ",
			elements:  []string{"one", "two", "three"},
			want:      "one, two, three",
		},
		{
			name:      "empty separator with empty elements filtered",
			separator: "",
			elements:  []string{"a", "", "b"},
			want:      "ab",
		},
		{
			name:      "only one non-empty among many empty",
			separator: "/",
			elements:  []string{"", "", "sole", "", ""},
			want:      "sole",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Join(tc.separator, tc.elements...)
			if got != tc.want {
				t.Errorf("Join(%q, %v) = %q, want %q", tc.separator, tc.elements, got, tc.want)
			}
		})
	}
}
