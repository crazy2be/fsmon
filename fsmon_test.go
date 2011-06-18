package fsmon

import (
	"testing"
	//s"time"
	"fmt"
	"log"
	"os"
)

type SomeHandler struct {
	t *testing.T
	modChan chan bool
	delChan chan bool
}

func NewSomeHandler(t *testing.T) *SomeHandler {
	sh := new(SomeHandler)
	sh.t = t
	sh.modChan = make(chan bool)
	sh.delChan = make(chan bool)
	return sh
}

func (sh *SomeHandler) Modified(name string) {
	sh.t.Log(name + " modified")
	sh.modChan <- true
	sh.t.Log(name + " modified event sent")
}

func (sh *SomeHandler) Deleted(name string) {
	sh.t.Log(name + " deleted")
	sh.delChan <- true
	sh.t.Log(name + " deleted event sent")
}

func init() {
	log.SetFlags(log.Lshortfile)
	f, err := os.Create("foo")
	if err != nil {
		log.Fatal(err.String() + " Before watch")
	}
	f.Close()/*
	go func() {
		time.Sleep(10*1000*1000*1000)
		panic("debug")
	}()*/
}

func setupHandler(t *testing.T) (*InotifyWatcher, *SomeHandler) {
	inw, err := NewInotifyWatcher()
	if err != nil {
		t.Fatal(err)
	}
	sh := NewSomeHandler(t)
	err = inw.AddWatch("foo", sh)
	if err != nil {
		t.Fatal(err)
	}
	sh.t = t
	go inw.Watch()
	return inw, sh
}

func modifyFile(t *testing.T, name string) {
	f, err := os.Create(name)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Fprintln(f, "Hello World")
	f.Close()
}

func TestModifiedHandler(t *testing.T) {
	inw, sh := setupHandler(t)
	
	inw.AddWatch("foo0", sh)
	modifyFile(t, "foo0")
	t.Log("Waiting for modified event 1...")
	t.Log(inw)
	<- sh.modChan
	<- sh.modChan
	
	modifyFile(t, "foo1")
	inw.AddWatch("foo1", sh)
	removeFile(t, "foo1")
	t.Log("Delete event...")
	t.Log(inw)
	select {
		case <- sh.modChan:
		case <- sh.delChan:
	}
	
	modifyFile(t, "foo2")
	inw.AddWatch("foo2", sh)
	removeFile(t, "foo2")
	t.Log("2...")
	t.Log(inw)
	select {
		case <- sh.modChan:
		case <- sh.delChan:
	}
	
	t.Log("SUCCESS!")
}

func removeFile(t *testing.T, name string) {
	err := os.Remove(name)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDeleteHandler(t *testing.T) {
	_, sh := setupHandler(t)
	removeFile(t, "foo")
	t.Log("Waiting for delete event...")
	<- sh.delChan
	t.Log("SUCCESS!")
}