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

var theWatcher Watcher

func init() {
	inw, err := NewInotifyWatcher()
	if err != nil {
		log.Fatal("fsmon failed to initialize any watchers :(.")
	}
	theWatcher = inw
}

