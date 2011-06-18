package fsmon

import (
	"os/inotify"
	"path/filepath"
)

type dirWatchHandler struct {
	path string
	fileWatchers []watchHandler
}

type watchHandler struct {
	path string
	handler Handler
}

type InotifyWatcher struct {
	watchHandlers []dirWatchHandler
	watcher *inotify.Watcher
}

func NewInotifyWatcher() (*InotifyWatcher, os.Error) {
	inw := new(InotifyWatcher)
	inw.watchHandlers = make([]dirWatchHandler)
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
}

// Removes all handlers for a given filename.
// BUG: Does not remove actual underlying watches, only removes the handlers from the list.
func (inw *InotifyWatcher) RemoveWatch(relname string) os.Error {
	name, err := filepath.Abs(relname)
	if err != nil {
		return err
	}
	watchHandlers[name] = nil, false
	rmDirWatch(name)
	return nil
}

// Blocks, watching for events forever and calling the appropriate callbacks.
func (inw *InotifyWatcher) Watch() {
	//defer theWatcher.Close()
	select {
		case ev := <-theWatcher.Event:
			callCallback(ev)
		case err := <-theWatcher.Error:
			log.Println(err)
	}
}

func (inw *InotifyWatcher) isWatchingDir(name string) bool {
	for _, dir := range watchedFolders {
		if dir == name {
			return true
		}
	}
	return false
}

func (inw *InotifyWatcher) callCallback(ev *inotify.Event) {
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

// Adds a watch to the directory for the file given by name.
func (inw *InotifyWatcher) addDirWatch(name string) os.Error {
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
func (inw *InotifyWatcher) rmDirWatch(name string) os.Error {
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