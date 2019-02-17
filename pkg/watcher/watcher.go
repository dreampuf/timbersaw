package watcher

import (
	"bufio"
	"context"
	"github.com/dreampuf/timbersaw/pkg/log"
	"github.com/fsnotify/fsnotify"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Watcher struct {
	paths  []string
	logger log.Logger
	ch     chan<- string

	processing sync.Map
}

func NewWatcher(logger log.Logger, paths string, ch chan<- string) *Watcher {
	return &Watcher{
		paths:      strings.Split(paths, ","),
		logger:     logger,
		processing: sync.Map{},
		ch:         ch,
	}
}

func (w *Watcher) Watch(ctx context.Context) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	var sg sync.WaitGroup

	go func() {
		w.logger.Debug("start watch")
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				if event.Op&fsnotify.Create == fsnotify.Create {
					if info, err := os.Stat(event.Name); err != nil {
						w.logger.Errorw("accessing watcher failed", "path", event.Name, "error", err)
					} else {
						if info.IsDir() {
							if err := watcher.Add(event.Name); err != nil {
								w.logger.Errorw("add watcher failed", "path", event.Name, "error", err)
							}
						} else {
							sg.Add(1)
							go w.monitorFile(ctx, &sg, event.Name, w.ch)
						}
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				w.logger.Error(err)

			case <-ctx.Done():
				w.logger.Debug("stop watch")
				return
			}
		}
	}()
	for _, path := range w.paths {
		w.walkthrough(ctx, watcher, &sg, path)
	}

	<-ctx.Done()
	sg.Wait()
	return nil
}

func (w *Watcher) walkthrough(ctx context.Context, watcher *fsnotify.Watcher, sg *sync.WaitGroup, path string) {
	accessingErr := func(path string, err error) {
		w.logger.Errorw("accessing file failed", "err", err, "path", path)
	}
	if fileInfo, err := os.Stat(path); err != nil {
		accessingErr(path, err)
	} else {
		if fileInfo.IsDir() {
			if err := watcher.Add(path); err != nil {
				accessingErr(path, err)
			}
			if err := filepath.Walk(path, func(spath string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if info.IsDir() {
					if err := watcher.Add(spath); err != nil {
						return err
					}
				} else {
					sg.Add(1)
					go w.monitorFile(ctx, sg, spath, w.ch)
				}
				return nil
			}); err != nil {
				accessingErr(path, err)
			}
		} else {
			sg.Add(1)
			go w.monitorFile(ctx, sg, path, w.ch)
		}
	}
}

func (w *Watcher) monitorFile(ctx context.Context, sg *sync.WaitGroup, path string, ch chan<- string) {
	defer sg.Done()

	if lastAccessTimestamp, ok := w.processing.Load(path); ok {
		w.logger.Warnw("duplicate access", "path", path, "access", lastAccessTimestamp)
		return
	} else {
		w.processing.Store(path, time.Now())
	}

	file, err := os.Open(path)
	if err != nil {
		w.logger.Errorw("open file failed", "error", err, "path", path)
		return
	}
	defer file.Close()

	func() {
		for {
			scanner := bufio.NewScanner(file)
			scanner.Split(bufio.ScanLines)

			for scanner.Scan() {
				ch <- scanner.Text()
			}

			if err := scanner.Err(); err != nil {
				w.logger.Errorw("scan file error", "error", err)
				return
			}

			//TODO a better idle schedule
			t := time.After(5 * time.Second)
			select {
			case <-ctx.Done():
				return
			case <-t:
			}
		}
	}()
}
