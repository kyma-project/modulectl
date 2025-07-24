package slices_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kyma-project/modulectl/internal/common/utils/slices"
)

func TestMergeAndDeduplicateSlices(t *testing.T) {
	tests := []struct {
		name     string
		input    [][]string
		expected []string
	}{
		{
			name:     "Overlapping elements",
			input:    [][]string{{"a", "b"}, {"b", "c"}},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "Unique elements",
			input:    [][]string{{"x"}, {"y"}, {"z"}},
			expected: []string{"x", "y", "z"},
		},
		{
			name:     "Empty strings ignored",
			input:    [][]string{{"", "a"}, {"b", ""}},
			expected: []string{"a", "b"},
		},
		{
			name:     "All slices empty",
			input:    [][]string{{}, {}},
			expected: []string{},
		},
		{
			name:     "No arguments",
			input:    [][]string{},
			expected: []string{},
		},
		{
			name:     "All duplicates",
			input:    [][]string{{"d", "d"}, {"d"}},
			expected: []string{"d"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := slices.MergeAndDeduplicate(tt.input...)
			require.ElementsMatch(t, tt.expected, got)
		})
	}
}

func TestSetToSlice(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]struct{}
		expected []string
	}{
		{
			name:     "Empty set",
			input:    map[string]struct{}{},
			expected: []string{},
		},
		{
			name:     "Single item",
			input:    map[string]struct{}{"a": {}},
			expected: []string{"a"},
		},
		{
			name:     "Multiple items",
			input:    map[string]struct{}{"a": {}, "b": {}, "c": {}},
			expected: []string{"a", "b", "c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := slices.SetToSlice(tt.input)
			require.ElementsMatch(t, tt.expected, got)
		})
	}
}
