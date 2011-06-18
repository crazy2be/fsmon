package fsmon

import (
	"testing"
	//s"time"
	"path"
	"fmt"
	"log"
	"os"
)

type SomeHandler struct {
	t *testing.T
	modChan chan bool
	delChan chan bool
	createChan chan bool
}

func NewSomeHandler(t *testing.T) *SomeHandler {
	sh := new(SomeHandler)
	sh.t = t
	sh.modChan = make(chan bool)
	sh.delChan = make(chan bool)
	sh.createChan = make(chan bool)
	return sh
}

func (sh *SomeHandler) Modified(name string) {
	tlog(sh.t, name + " modified")
	sh.modChan <- true
	tlog(sh.t, name + " modified event sent")
}

func (sh *SomeHandler) Deleted(name string) {
	tlog(sh.t, name + " deleted")
	sh.delChan <- true
	tlog(sh.t, name + " deleted event sent")
}

func (sh *SomeHandler) Created(name string) {
	tlog(sh.t, name + " created")
	sh.createChan <- true
	tlog(sh.t, name + " created event sent")
}

const (
	// gotest only prints the results of t.Log() after the test is completed. However, the tests here often hang when they fail, so you can set this to true to see the output as it comes.
	DEBUG_HANG = false
)

func tlog(t *testing.T, i ...interface{}) {
	if DEBUG_HANG {
		log.Println(i...)
	} else {
		t.Log(i...)
	}
}

func init() {
	log.SetFlags(log.Lshortfile)
	
	err := os.RemoveAll("/tmp/_tests_3932")
	if err != nil {
		log.Fatal(err)
	}
	err = os.Mkdir("/tmp/_tests_3932", 0777)
	if err != nil {
		log.Fatal(err)
	}
}

func setupHandler(t *testing.T) (Watcher, *SomeHandler) {
	inw, err := NewWatcher()
	if err != nil {
		t.Fatal(err)
	}
	sh := NewSomeHandler(t)
	go inw.Watch()
	return inw, sh
}

func createFile(t *testing.T, name string) {
	f, err := os.Create(name)
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
}

func modifyFile(t *testing.T, name string) {
	f, err := os.Create(name)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Fprintln(f, "Hello World")
	f.Close()
}

func deleteFile(t *testing.T, name string) {
	err := os.Remove(name)
	if err != nil {
		t.Fatal(err)
	}
}

func testfile(name string) string {
	return path.Join("/tmp/_tests_3932", name)
}

func TestMultipleHandlers(t *testing.T) {
	tlog(t, "Setting up handler...")
	inw, sh := setupHandler(t)
	
	tlog(t, "Adding first watch...")
	createFile(t, testfile("foo0"))
	inw.AddWatch(testfile("foo0"), sh)
	modifyFile(t, testfile("foo0"))
	tlog(t, "Waiting for modified event 1...")
	tlog(t, inw)
	<- sh.modChan
	<- sh.modChan
	
	modifyFile(t, testfile("foo1"))
	inw.AddWatch(testfile("foo1"), sh)
	deleteFile(t, testfile("foo1"))
	tlog(t, "Delete event...")
	tlog(t, inw)
	// Unspecified order
	select {
		case <- sh.createChan:
		case <- sh.modChan:
		case <- sh.delChan:
	}
	
	modifyFile(t, testfile("foo2"))
	inw.AddWatch(testfile("foo2"), sh)
	deleteFile(t, testfile("foo2"))
	tlog(t, "2...")
	tlog(t, inw)
	// Unspecified order
	select {
		case <- sh.createChan:
		case <- sh.modChan:
		case <- sh.delChan:
	}
	
	tlog(t, "SUCCESS!")
}

func TestModifiedHandler(t *testing.T) {
	createFile(t, testfile("foo"))
	inw, sh := setupHandler(t)
	
	err := inw.AddWatch(testfile("foo"), sh)
	if err != nil {
		t.Fatal(err)
	}
	modifyFile(t, testfile("foo"))
	tlog(t, "Waiting for modified event...")
	<- sh.modChan
	
	tlog(t, "SUCCESS!")
}

func TestDeleteHandler(t *testing.T) {
	createFile(t, testfile("foo"))
	inw, sh := setupHandler(t)
	
	err := inw.AddWatch(testfile("foo"), sh)
	if err != nil {
		t.Fatal(err)
	}
	
	deleteFile(t, testfile("foo"))
	tlog(t, "Waiting for delete event...")
	<- sh.delChan
	tlog(t, "SUCCESS!")
}

func TestFolderHandler(t *testing.T) {
	inw, sh := setupHandler(t)
	
	err := os.Mkdir(testfile("foodir1"), 0777)
	if err != nil {
		t.Fatal(err)
	}
	createFile(t, testfile("foodir1/foo"))
	err = inw.AddWatch(testfile("foodir1"), sh)
	
	modifyFile(t, testfile("foodir1/foo"))
	tlog(t, "Waiting for modified event...")
	<- sh.modChan
	<- sh.modChan
	
	createFile(t, testfile("foodir1/foo2"))
	modifyFile(t, testfile("foodir1/foo2"))
	tlog(t, "Waiting for created event...")
	select {
		case <- sh.createChan:
		case <- sh.modChan:
	}
}