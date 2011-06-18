package fsmon

import (
	"log"
	"os"
	"os/inotify"
	"path"
	"path/filepath"
)

type watchHandler struct {
	path string
	handler Handler
}

type InotifyWatcher struct {
	watchHandlers []watchHandler
	watcher *inotify.Watcher
}

func NewInotifyWatcher() (*InotifyWatcher, os.Error) {
	inw := new(InotifyWatcher)
	inw.watchHandlers = make([]watchHandler, 0)
	var err os.Error
	inw.watcher, err = inotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return inw, nil
}

// Adds a handler that will be called for events on the file given by name.
func (inw *InotifyWatcher) AddWatch(name string, h Handler) os.Error {
	var err os.Error
	name, err = filepath.Abs(name)
	if err != nil {
		return err
	}
	
	err = inw.addDirWatch(name)
	if err != nil {
		return err
	}
	
	var watcher watchHandler
	
	watcher.path = name
	watcher.handler = h
	
	inw.watchHandlers = append(inw.watchHandlers, watcher)
	
	return nil
}

// Removes all handlers for a given filename.
func (inw *InotifyWatcher) RemoveWatches(name string) os.Error {
	name, err := filepath.Abs(name)
	if err != nil {
		return err
	}
	
	i := inw.watchEntryWithName(name)
	for i > 0 {
		inw.removeWatchEntry(i)
		i = inw.watchEntryWithName(name)
	}
	
	inw.removeDirWatch(name)
	
	return nil
}

func (inw *InotifyWatcher) watchEntryWithName(name string) int {
	for index, entry := range inw.watchHandlers {
		if entry.path == name {
			return index
		}
	}
	return -1
}

func (inw *InotifyWatcher) removeWatchEntry(index int) {
	newWatchHandlers := make([]watchHandler, len(inw.watchHandlers)-1)
	for j := 0; j < index; j++ {
		newWatchHandlers[j] = inw.watchHandlers[j]
	}
	for j := index+1; j < len(newWatchHandlers); j++ {
		newWatchHandlers[j] = inw.watchHandlers[j+1]
	}
	inw.watchHandlers = newWatchHandlers
}

// Blocks, watching for events forever and calling the appropriate callbacks.
func (inw *InotifyWatcher) Watch() {
	defer inw.watcher.Close()
	for {
		select {
			case ev := <-inw.watcher.Event:
				inw.handleCallbacks(ev)
			case err := <-inw.watcher.Error:
				log.Println(err)
		}
	}
}

func (inw *InotifyWatcher) handleCallbacks(ev *inotify.Event) {
	name := path.Clean(ev.Name)
	for _, watch := range inw.watchHandlers {
		if name == watch.path {
			inw.handleCallback(ev, name, &watch)
		} else if dir, _ := getDir(name); watch.path == dir {
			inw.handleCallback(ev, name, &watch)
		}
	}
}

func (inw *InotifyWatcher) handleCallback(ev *inotify.Event, name string, wh *watchHandler) {
	m := ev.Mask
	switch m {
		case inotify.IN_MODIFY:
			c, ok := wh.handler.(ModifiedHandler)
			if ok {
				//log.Println(ev)
				c.Modified(name)
			}
		case inotify.IN_DELETE:
			c, ok := wh.handler.(DeletedHandler)
			if ok {
				//log.Println(ev)
				c.Deleted(name)
			}
		case inotify.IN_CREATE:
			c, ok := wh.handler.(CreatedHandler)
			if ok {
				c.Created(name)
			}
		default:
			if
				inotify.IN_OPEN & m != 0 ||
				m & inotify.IN_CLOSE != 0 ||
				m & inotify.IN_ACCESS != 0 {
					return
			}
			log.Println("Warning: Unhandled inotify event", ev)
	}
}

// Goes through the list of handlers to find out if we are currently watching the given directory (that is, there is a entry for it in the list)
func (inw *InotifyWatcher) isWatchingDir(name string) bool {
	for _, watchHandler := range inw.watchHandlers {
		hname, err := getDir(watchHandler.path)
		if err != nil {
			return false
		}
		if hname == name {
			return true
		}
	}
	return false
}

// Actually adds a watch to the directory for the file given by name on the underlying inotify object, if one does not exist already. Because inotify can watch whole directories, and there is a limit on the number of items one can watch at a time, we always watch directories rather than individual files.
func (inw *InotifyWatcher) addDirWatch(name string) os.Error {
	dirname, err := getDir(name)
	if err != nil {
		return err
	}
	// Are we already watching this directory?
	if inw.isWatchingDir(dirname) {
		return nil
	}
	
	err = inw.watcher.Watch(dirname)
	if err != nil {
		return err
	}
	
	return nil
}

// Removes a directory watch for the file given by name if no other handlers are in this directory.
func (inw *InotifyWatcher) removeDirWatch(name string) os.Error {
	dirname, err := getDir(name)
	if err != nil {
		return err
	}
	
	if !inw.isWatchingDir(dirname) {
		return nil
	}
	
	for _, watcher := range inw.watchHandlers {
		if dirname[:] == watcher.path[:len(dirname)] {
			// There's still a handler for this folder.
			return nil
		}
	}
	
	err = inw.watcher.RemoveWatch(dirname)
	if err != nil {
		return err
	}
	
	return nil
}