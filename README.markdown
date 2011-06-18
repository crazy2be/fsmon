Filesystem Change Monitor
=========================

Getting Started
---------------
Install it:

	goinstall github.com/crazy2be/fsmon

Import it:

	import "github.com/crazy2be/fsmon"

Use it:

	fsmon.AddWatch("foo", fooHandler)

Introduction
------------
The `fsmon` library provides a callback-based API for monitoring changes to files in the filesystem. It is designed to be operating system-agnostic, although it currently only supports inotify as a backend.

Adding Support For Other APIs
-----------------------------
Each different watcher type (`InotifyWatcher`, `Win32APIWatcher`, etc) needs only satisfy the Watcher interface, which is fairly simple:

	type Watcher interface {
		AddWatch(string, Handler) os.Error
		RemoveWatches(string) os.Error
		Watch()
	}

`AddWatch()` adds another `Handler` for the path specified. If a handler already exists on that path, it is not affected. Adding a handler to a file gives notifications only for that file, adding a handler to a directory gives all notifications for the immediate desendents of that directory. Directory watches are not recursive. The underlying watcher can "batch" handlers for multiple files in the same directory if it is more efficent. For example, the `InotifyWatcher` only watches directories rather than files on the underlying watcher, but makes this optimization transparent to users.

`RemoveWatches()` removes all `Handler`s for the path specified. As far as I can see, there is no way to allow someone to remove a specific callback, so this removes all of them.

`Watch()` blocks forever in a loop that listens for events and calls the relevent callbacks. No callbacks will be called before Watch() is called. `AddWatch()` and `RemoveWatches()` should be callable after `Watch()`, and should not be subject to race conditions.