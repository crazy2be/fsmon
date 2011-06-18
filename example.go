package main

import (
	"log"
	"github.com/crazy2be/fsmon"
)

type FooHandler int

func (fh *FooHandler) Modified(name string) {
	log.Println(name + " was modified")
}

func (fh *FooHandler) Created(name string) {
	log.Println(name + " was created")
}

func (fh *FooHandler) Deleted(name string) {
	log.Println(name + " was deleted")
}

func main() {
	fh := new(FooHandler)
	
	fsmon.AddWatch("foo", fh)
	log.Println("Listening for events in folder or affecting file 'foo'")
	
	fsmon.Watch()
}
