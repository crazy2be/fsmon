package fsmon

import (
	"testing"
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
	sh.modChan = make(chan bool, 1)
	sh.delChan = make(chan bool, 1)
	return sh
}

func (sh *SomeHandler) Modified(name string) {
	fmt.Println(name + " modified")
	sh.modChan <- true
}

func (sh *SomeHandler) Deleted(name string) {
	fmt.Println(name + " deleted")
	sh.delChan <- true
}

func init() {
	log.SetFlags(log.Lshortfile)
	f, err := os.Create("foo")
	if err != nil {
		log.Fatal(err.String() + " Before watch")
	}
	f.Close()
}

func setupHandler(t *testing.T) *SomeHandler {
	inw, err := NewInotifyWatcher()
	if err != nil {
		t.Fatal(err)
	}
	sh := NewSomeHandler(t)
	err = inw.AddWatch("foo", sh)
	if err != nil {
		log.Panic(err)
		t.Fatal(err)
	}
	sh.t = t
	go inw.Watch()
	return sh
}

func TestModifiedHandler(t *testing.T) {
	sh := setupHandler(t)
	f, err := os.Create("foo")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Fprintln(f, "Hello World")
	f.Close()
	log.Println("Waiting for modified event...")
	<- sh.modChan
	log.Println("SUCCESS!")
}

func TestDeleteHandler(t *testing.T) {
	sh := setupHandler(t)
	err := os.Remove("foo")
	if err != nil {
		t.Fatal(err)
	}
	log.Println("Waiting for delete event...")
	<- sh.delChan
	log.Println("SUCCESS!")
}