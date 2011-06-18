// Implements os-independent callbacks for file modification events. Currently only supports inotify (linux)
package fsmon

import (
	"os"
	"log"
	"path"
	"path/filepath"
	"os/inotify"
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

func isDir(name string) (bool, os.Error) {
	finfo, err := os.Stat(name)
	if err != nil {
		return false, err
	}
	if finfo.IsDirectory() {
		return true, nil
	}
	return false, nil
}

func getDir(name string) (string, os.Error) {
	name, err := filepath.Abs(name)
	if err != nil {
		return "", err
	}
	isdir, err := isDir(name)
	if err != nil {
		return "", err
	}
	if isdir {
		return name, nil
	}
	dirname, _ := path.Split(name)
	// Is this second check necessary?
	isdir, err = isDir(dirname)
	if err != nil {
		return "", err
	}
	if !isdir {
		return "", os.NewError("Path '"+name+"' is not a directory, and path package was unable to find the parent directory properly (tried '"+dirname+"')")
	}
	return dirname, nil
}