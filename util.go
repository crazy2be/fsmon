package fsmon

import (
	"os"
	"path"
	"path/filepath"
)

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
	dirname = path.Clean(dirname)
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