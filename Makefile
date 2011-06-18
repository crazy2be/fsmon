include $(GOROOT)/src/Make.inc

TARG=fsmon
GOFILES=\
	fsmon.go\
	handler.go\
	inotify.go\
	util.go

include $(GOROOT)/src/Make.pkg