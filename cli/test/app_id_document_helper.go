package test

import (
	"os"
	"testing"

	"numerous/cli/tool"

	"github.com/stretchr/testify/require"
)

const appIDFilePerm = 0o644

func ChdirToTmpDirWithAppIDDocument(t *testing.T, id string) {
	t.Helper()
	tmpDir := t.TempDir()
	require.NoError(t, os.Chdir(tmpDir))
	require.NoError(t, os.WriteFile(tool.ToolIDFileName, []byte(id), appIDFilePerm))
}
