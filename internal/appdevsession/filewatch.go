package appdevsession

import (
	"log/slog"
	"time"

	"github.com/fsnotify/fsnotify"
)

type clock interface {
	Now() time.Time
	Since(time.Time) time.Duration
}

type timeclock struct{}

func (c *timeclock) Now() time.Time {
	return time.Now()
}

func (c *timeclock) Since(t time.Time) time.Duration {
	return time.Since(t)
}

type fileWatcherFactory interface {
	// Create a FileWatcher
	Create() (FileWatcher, error)
}

type FSNotifyFileWatcherFactory struct{}

func (f *FSNotifyFileWatcherFactory) Create() (FileWatcher, error) {
	if watcher, err := fsnotify.NewBufferedWatcher(1); err != nil {
		return nil, err
	} else {
		return &FSNotifyWatcher{watcher: watcher}, nil
	}
}

type FileWatcher interface {
	// Close the file watcher
	Close() error
	// Add a file to the file watcher
	Add(name string) error
	// Get a channel of file events noticed by the FileWatcher
	GetEvents() chan fsnotify.Event
	// Get a channel of errors for the FileWatcher
	GetErrors() chan error
}

type FSNotifyWatcher struct {
	watcher *fsnotify.Watcher
}

func (w *FSNotifyWatcher) Close() error {
	return w.watcher.Close()
}

func (w *FSNotifyWatcher) Add(name string) error {
	return w.watcher.Add(name)
}

func (w *FSNotifyWatcher) GetEvents() chan fsnotify.Event {
	return w.watcher.Events
}

func (w *FSNotifyWatcher) GetErrors() chan error {
	return w.watcher.Errors
}

// Watch the given path for file changes, returning a channel to which is sent
// the file names of updated files.
//
// Updates happening within minInterval time of the last update are ignored.
func WatchAppChanges(path string, watcherFactory fileWatcherFactory, clock clock, minInterval time.Duration) (chan string, error) {
	updates := make(chan string)

	watcher, err := watcherFactory.Create()
	if err != nil {
		return nil, err
	}

	err = watcher.Add(path)
	if err != nil {
		watcher.Close()
		return nil, err
	}

	go func() {
		defer close(updates)
		defer watcher.Close()
		lastUpdate := clock.Now()
		for {
			select {
			case event, ok := <-watcher.GetEvents():
				if !ok {
					return
				}
				elapsed := clock.Since(lastUpdate)

				if event.Has(fsnotify.Write) && elapsed > minInterval {
					lastUpdate = clock.Now()
					updates <- event.Name
				}
			case err, ok := <-watcher.GetErrors():
				if !ok {
					return
				}
				slog.Warn("error watching app code", slog.Any("error", err))
			}
		}
	}()

	return updates, nil
}
