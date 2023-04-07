package strings_test

import (
	"testing"

	"github.com/facebookincubator/ttpforge/pkg/strings"
)

func TestStringSlicesEqual(t *testing.T) {
	testCases := []struct {
		name     string
		a        []string
		b        []string
		expected bool
	}{
		{
			name:     "Equal slices",
			a:        []string{"apple", "banana", "cherry"},
			b:        []string{"apple", "banana", "cherry"},
			expected: true,
		},
		{
			name:     "Unequal slices - different values",
			a:        []string{"apple", "banana", "cherry"},
			b:        []string{"apple", "banana", "grape"},
			expected: false,
		},
		{
			name:     "Unequal slices - different lengths",
			a:        []string{"apple", "banana", "cherry"},
			b:        []string{"apple", "banana"},
			expected: false,
		},
		{
			name:     "Empty slices",
			a:        []string{},
			b:        []string{},
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := strings.StringSlicesEqual(tc.a, tc.b)
			if result != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, result)
			}
		})
	}
}
