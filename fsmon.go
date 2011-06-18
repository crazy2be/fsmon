// Implements os-independent callbacks for file modification events. Currently only supports inotify (linux)
package fsmon

import (
	"os"
	"log"
	"path"
	"path/filepath"
	"os/inotify"
)

type Handler interface {
	
}

type ModifiedHandler interface {
	Handler
	Modified(name string)
}

type DeletedHandler interface {
	Handler
	Deleted(name string)
}

// TODO, not implemented
type MovedHandler interface {
	Handler
	Moved(source string, dest string)
}

var watchHandlers map[string][]Handler
var watchedFolders []string
var theWatcher *inotify.Watcher

func init() {
	watchHandlers = make(map[string][]Handler, 1)
	var err os.Error
	theWatcher, err = inotify.NewWatcher()
	if err != nil {
		log.Panicln("Failed to initilize inotify! fsmon is unhappy :(")
	}
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

// Adds a watch to the directory for the file given by name.
func addDirWatch(name string) os.Error {
	dirname, err := getDir(name)
	if err != nil {
		return err
	}
	// Are we already watching this directory?
	if isWatchingDir(dirname) {
		return nil
	}
	
	err = theWatcher.Watch(dirname)
	if err != nil {
		return err
	}
	watchedFolders = append(watchedFolders, dirname)
	return nil
}

// Removes a directory watch for the file given by name if no other handlers are in this directory.
func rmDirWatch(name string) os.Error {
	dirname, err := getDir(name)
	if err != nil {
		return err
	}
	if !isWatchingDir(dirname) {
		return nil
	}
	for path, _ := range watchHandlers {
		if dirname[:] == path[:len(dirname)] {
			// There's still a handler for this folder.
			return nil
		}
	}
	for i, dir := range watchedFolders {
		if dir == dirname {
			newWatchedFolders := make([]string, len(watchedFolders)-1)
			for j := 0; j < i; j++ {
				newWatchedFolders[j] = watchedFolders[j]
			}
			for j := i+1; j < len(newWatchedFolders); j++ {
				newWatchedFolders[j] = watchedFolders[j+1]
			}
			watchedFolders = newWatchedFolders
			err = theWatcher.RemoveWatch(dirname)
			if err != nil {
				return err
			}
			return nil
		}
	}
	return nil
}

func isWatchingDir(name string) bool {
	for _, dir := range watchedFolders {
		if dir == name {
			return true
		}
	}
	return false
}

// Adds a handler that will be called for events on the file given by relname.
func AddHandler(h Handler, relname string) os.Error {
	name, err := filepath.Abs(relname)
	if err != nil {
		return err
	}
	
	err = addDirWatch(name)
	if err != nil {
		return err
	}
	
	handlers := watchHandlers[name]
	if handlers == nil {
		handlers = make([]Handler, 1)
		handlers[0] = h
	} else {
		handlers = append(handlers, h)
	}
	watchHandlers[name] = handlers
	return nil
}

// Removes all handlers for a given filename.
// BUG: Does not remove actual underlying watches, only removes the handlers from the list.
func RemoveHandlers(relname string) os.Error {
	name, err := filepath.Abs(relname)
	if err != nil {
		return err
	}
	watchHandlers[name] = nil, false
	rmDirWatch(name)
	return nil
}

func callCallback(ev *inotify.Event) {
	log.Println(ev)
	log.Println(watchHandlers)
	log.Println(watchedFolders)
	name := path.Clean(ev.Name)
	callbacks, ok := watchHandlers[name]
	// No handlers exist
	if !ok {
		return
	}
	for _, callback := range callbacks {
		m := ev.Mask
		switch m {
			case inotify.IN_MODIFY:
				c, ok := callback.(ModifiedHandler)
				if !ok {
					continue;
				}
				c.Modified(name)
			case inotify.IN_DELETE:
				c, ok := callback.(DeletedHandler)
				if !ok {
					continue
				}
				c.Deleted(name)
			default:
				log.Println("Warning: Unhandled inotify event", ev)
		}
	}
}

// Blocks, watching for events forever and calling the appropriate callbacks.
func Watch() {
	//defer theWatcher.Close()
	select {
		case ev := <-theWatcher.Event:
			callCallback(ev)
		case err := <-theWatcher.Error:
			log.Println(err)
	}
}