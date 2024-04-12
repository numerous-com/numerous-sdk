package initialize

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractPythonVersion(t *testing.T) {
	tests := []struct {
		version                  string
		expectedExtractedVersion string
	}{
		{
			version:                  "3.12.4",
			expectedExtractedVersion: "3.12",
		},
		{
			version:                  "3.7.0",
			expectedExtractedVersion: "3.7",
		},
		{
			version:                  "2.7",
			expectedExtractedVersion: "2.7",
		},
		{
			version:                  "1.5.1p1",
			expectedExtractedVersion: "1.5",
		},
	}
	for _, test := range tests {
		t.Run("Can extract python version "+test.expectedExtractedVersion, func(t *testing.T) {
			actualExtractedVersion, err := extractPythonVersion([]byte("Python " + test.version))
			require.NoError(t, err)
			assert.Equalf(t, test.expectedExtractedVersion, actualExtractedVersion, test.expectedExtractedVersion+"=="+actualExtractedVersion)
		})
	}
}
