package test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

const appIDFilePerm = 0o644

func ChdirToTmpDirWithAppIDDocument(t *testing.T, appIDFile, id string) {
	t.Helper()
	tmpDir := t.TempDir()
	require.NoError(t, os.Chdir(tmpDir))
	require.NoError(t, os.WriteFile(appIDFile, []byte(id), appIDFilePerm))
}
