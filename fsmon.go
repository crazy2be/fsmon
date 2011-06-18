// Implements os-independent callbacks for file modification events. Currently only supports inotify (linux)
package fsmon

import (
	"os"
	"log"
)

type Watcher interface {
	AddWatch(string, Handler) os.Error
	RemoveWatches(string) os.Error
	Watch()
}

var defaultWatcher Watcher

func init() {
	var err os.Error
	defaultWatcher, err = NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
}

func AddWatch(name string, h Handler) os.Error {
	return defaultWatcher.AddWatch(name, h)
}

func RemoveWatches(name string) os.Error {
	return defaultWatcher.RemoveWatches(name)
}

func Watch() {
	defaultWatcher.Watch()
}

func NewWatcher() (Watcher, os.Error) {
	inw, err := NewInotifyWatcher()
	if err == nil {
		return inw, nil
	}
	// TODO: Add more watcher fallbacks here for different operating systems.
	return os.NewError("fsmon failed to initialize any watchers :(.")
}