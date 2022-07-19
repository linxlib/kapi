package rundaemon

import (
	"errors"
	"fmt"
	"github.com/linxlib/kapi/cmd/k/utils/innerlog"
	pollingWatcher "github.com/radovskyb/watcher"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

func directoryShouldBeTracked(cfg *WatcherConfig, path string) bool {
	return cfg.flagRecursive == true && !cfg.flagExcludedDirs.Matches(path)
}

func pathMatches(cfg *WatcherConfig, path string) bool {
	base := filepath.Base(path)
	return (cfg.flagIncludedFiles.Matches(base) || matchesPattern(cfg.pattern, path)) &&
		!cfg.flagExcludedFiles.Matches(base)
}

type WatcherConfig struct {
	flagVerbose       bool
	flagRecursive     bool
	flagDirectories   globList
	flagExcludedDirs  globList
	flagExcludedFiles globList
	flagIncludedFiles globList
	pattern           *regexp.Regexp
}

type FileWatcher interface {
	Close() error
	AddFiles() error
	add(path string) error
	Watch(jobs chan<- string)
	getConfig() *WatcherConfig
}

type PollingWatcher struct {
	watcher *pollingWatcher.Watcher
	cfg     *WatcherConfig
}

func (p PollingWatcher) Close() error {
	p.watcher.Close()
	return nil
}

func (p PollingWatcher) AddFiles() error {
	p.watcher.AddFilterHook(pollingWatcher.RegexFilterHook(p.cfg.pattern, false))

	return addFiles(p)
}

func (p PollingWatcher) Watch(jobs chan<- string) {
	// Start the watching process.
	go func() {
		if err := p.watcher.Start(time.Duration(100) * time.Millisecond); err != nil {
			innerlog.Log.Fatalln(err)
		}
	}()

	for {
		select {
		case event := <-p.watcher.Event:
			if p.cfg.flagVerbose {
				// Print the event's info.
				innerlog.Log.Println(event)
			}

			if pathMatches(p.cfg, event.Path) {
				jobs <- event.String()
			}
		case err := <-p.watcher.Error:
			if err == pollingWatcher.ErrWatchedFileDeleted {
				continue
			}
			innerlog.Log.Fatalln(err)
		case <-p.watcher.Closed:
			return
		}
	}
}

func (p PollingWatcher) add(path string) error {
	return p.watcher.Add(path)
}

func (p PollingWatcher) getConfig() *WatcherConfig {
	return p.cfg
}

func NewWatcher(cfg *WatcherConfig) (FileWatcher, error) {
	if cfg == nil {
		err := errors.New("no config specified")
		return nil, err
	}
	w := pollingWatcher.New()
	return PollingWatcher{
		watcher: w,
		cfg:     cfg,
	}, nil
}

func addFiles(fw FileWatcher) error {
	cfg := fw.getConfig()
	for _, flagDirectory := range cfg.flagDirectories {
		if cfg.flagRecursive == true {
			err := filepath.Walk(flagDirectory, func(path string, info os.FileInfo, err error) error {
				if err == nil && info.IsDir() {
					if cfg.flagExcludedDirs.Matches(path) {
						return filepath.SkipDir
					} else {
						if cfg.flagVerbose {
							innerlog.Log.Printf("Watching directory '%s' for changes.\n", path)
						}
						return fw.add(path)
					}
				}
				return err
			})

			if err != nil {
				return fmt.Errorf("filepath.Walk(): %v", err)
			}

			if err := fw.add(flagDirectory); err != nil {
				return fmt.Errorf("FileWatcher.Add(): %v", err)
			}
		} else {
			if err := fw.add(flagDirectory); err != nil {
				return fmt.Errorf("FileWatcher.AddFiles(): %v", err)
			}
		}
	}
	return nil
}
