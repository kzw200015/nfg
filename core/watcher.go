package core

import (
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/kzw200015/nfg/config"
	"github.com/kzw200015/nfg/log"
)

const defaultInterval = time.Duration(30) * time.Second

type Watcher struct {
	src  string
	temp string
	mu   sync.Mutex
}

func NewWatcher(src, temp string) *Watcher {
	return &Watcher{src, temp, sync.Mutex{}}
}

func (w *Watcher) Watch() {
	var fsChan <-chan error
	if !config.IsRemote(w.src) {
		fsChan = w.watchFs()
	}

	schChan := w.schedule()

	for {
		select {
		case err := <-fsChan:
			log.Logger.Errorln(err)
		case err := <-schChan:
			log.Logger.Errorln(err)
		}
	}
}

func (w *Watcher) schedule() <-chan error {
	errChan := make(chan error)
	go func() {
		for {
			go w.do(errChan)
			time.Sleep(defaultInterval)
		}
	}()
	return errChan
}

func (w *Watcher) watchFs() <-chan error {
	errChan := make(chan error)

	go func() {
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			log.Logger.Panicln(err)
		}

		defer watcher.Close()

		err = watcher.Add(w.src)
		if err != nil {
			log.Logger.Panicln(err)
		}
		for {
			select {
			case event := <-watcher.Events:
				if (event.Op == fsnotify.Write) || (event.Op == fsnotify.Chmod) {
					go w.do(errChan)
				}

				if event.Op == fsnotify.Remove {
					err = watcher.Add(w.src)
					if err != nil {
						errChan <- err
					}
				}
			case err := <-watcher.Errors:
				if err != nil {
					errChan <- err
				}
			}
		}
	}()

	return errChan
}

func (w *Watcher) do(errChan chan<- error) {
	w.mu.Lock()
	c, err := config.NewLocalConfig(w.src)
	if err != nil {
		errChan <- err
	}
	if err = SaveToFile(c.Rules, w.temp); err != nil {
		errChan <- err
	}
	if err = Apply(w.temp); err != nil {
		errChan <- err
	}
	w.mu.Unlock()
}
