package push

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShouldExclude(t *testing.T) {
	tests := []struct {
		excludePatterns []string
		includedFiles   []string
		excludedFiles   []string
	}{
		{
			excludePatterns: []string{"*.go"},
			excludedFiles:   []string{"exclude_file.go", "included_folder/excluded_file.go"},
			includedFiles:   []string{"included_file.py", "included_folder/included_file.py"},
		},
		{
			excludePatterns: []string{"venv"},
			excludedFiles:   []string{"venv/exclude_file.py", "venv/folder/excluded_file.py"},
			includedFiles:   []string{"included_file.py", "included_folder/included_file.py"},
		},
		{
			excludePatterns: []string{"*venv"},
			excludedFiles:   []string{".venv/exclude_file.py", ".venv/folder/excluded_file.py", "venv/excluded_file.py"},
			includedFiles:   []string{"included_file.py", "included_folder/included_file.py"},
		},
		{
			excludePatterns: []string{"venv*"},
			excludedFiles:   []string{"venv1/exclude_file.py", "venv1/folder/excluded_file.py", "venv/excluded_file.py"},
			includedFiles:   []string{"included_file.py", "1venv/included_file.py"},
		},
		{
			excludePatterns: []string{"venv"},
			excludedFiles:   []string{"venv/exclude_file.py", "venv/folder/excluded_file.py"},
			includedFiles:   []string{"included_file.py", "included_folder/included_file.py"},
		},
		{
			excludePatterns: []string{"excluded_file.py"},
			excludedFiles:   []string{"venv/excluded_file.py", "included_folder/excluded_file.py", "excluded_file.py"},
			includedFiles:   []string{"included_file.py", "included_folder/included_file.py"},
		},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("Testing pattern: %v", test.excludePatterns), func(t *testing.T) {
			for _, path := range test.includedFiles {
				assert.False(t, shouldExclude(test.excludePatterns, path), path)
			}
			for _, path := range test.excludedFiles {
				assert.True(t, shouldExclude(test.excludePatterns, path), path)
			}
		})
	}
}
