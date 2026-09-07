package syspower

import (
	"os"
	"testing"
)

func TestWatchedFile(t *testing.T) {
	f, _ := os.CreateTemp(os.TempDir(), "watched_file_test")
	fileName := f.Name()
	f.Close()

	os.WriteFile(fileName, []byte("1"), 0600)

	isChanged := false

	fw := NewWatchedFile(fileName, func(oldValue, newValue string) {
		isChanged = true
	})

	if fw.Value() != "1" {
		t.Fatalf("TestWatchedFile(): Expected initial value was '1', but is %s", fw.Value())
	}

	fw.Write("0")

	fw.Read()

	if !isChanged {
		t.Fatal("TestWatchedFile(): On change event wasn't invoked!")
	}

	defer os.Remove(fileName)
}
