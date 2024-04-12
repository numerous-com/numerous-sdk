package appdevsession

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWatchToolChanges(t *testing.T) {
	t.Run("sends message on file write", func(t *testing.T) {
		dir := t.TempDir()
		filepath := dir + "/file.txt"

		err := os.WriteFile(filepath, []byte("content before update\n"), 0o700)
		require.NoError(t, err)

		changes, err := WatchAppChanges(filepath, &FSNotifyFileWatcherFactory{}, &timeclock{}, time.Second*0)
		require.NoError(t, err)

		//nolint:errcheck
		go os.WriteFile(filepath, []byte("content after update 5\n"), 0o700)

		select {
		case changedFile := <-changes:
			assert.Equal(t, filepath, changedFile)
		case <-time.After(time.Second * 4):
			assert.FailNow(t, "timed out waiting for file update")
		}
	})
}
