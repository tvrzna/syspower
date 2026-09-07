package syspower

import (
	"os"
	"slices"
	"testing"
)

func TestSysFsVal(t *testing.T) {
	f, _ := os.CreateTemp(os.TempDir(), "watched_file_test")
	fileName := f.Name()
	f.Close()

	os.WriteFile(fileName, []byte("1"), 0600)

	fw := NewWatchedFile(fileName, nil)

	s := NewStaticFsVal(fw, []string{"0", "1"})

	if !slices.Contains(s.Choices(), "1") || !slices.Contains(s.Choices(), "0") || len(s.Choices()) != 2 {
		t.Fatalf("TestSysFsVal(): Unexpected slice of choices: %v", s.Choices())
	}

	if val := s.Value(); val != "1" {
		t.Fatalf("TestSysFsVal(): Unexpected value, expected 1, got %s", val)
	}

	if s.Set("2") == nil {
		t.Fatal("TestSysFsVal(): Setting value out of choices didn't led to error")
	}

	if err := s.Set("0"); err != nil {
		t.Fatalf("TestSysFsVal(): Setting value out of choices led to error: %v", err)
	}

	if s.Value() == "0" {
		t.Fatal("TestSysFsVal(): New value shouldn't be loaded yet")
	}

	if err := s.UpdateValue(); err != nil {
		t.Fatalf("TestSysFsVal(): Updating value of watched file led to error: %v", err)
	}

	if val := s.Value(); val != "0" {
		t.Fatalf("TestSysFsVal(): value after change is incorrect, expected 0, got %s", val)
	}
}
