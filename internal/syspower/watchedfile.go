package syspower

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"sync"
)

type WatchedFile struct {
	onChange OnChange
	path     string
	value    string
	mu       sync.RWMutex
}

type OnChange func(watchedFile *WatchedFile)

// Creates new watched file
func NewWatchedFile(path string, onChange OnChange) *WatchedFile {
	result := &WatchedFile{}
	result.path = path
	result.onChange = onChange
	result.readInternal(false)

	return result
}

// Gets lastly loaded loaded from watched file.
func (w *WatchedFile) Value() string {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.value
}

// Reads the watched file and invokes onChange event, if is defined and the content of file has changed.
func (w *WatchedFile) Read() error {
	return w.readInternal(true)
}

// Reads the watched file and invokes onChange event, if is defined and the content of file has changed.
// For internal usage only.
func (w *WatchedFile) readInternal(notify bool) error {
	data, err := os.ReadFile(w.path)
	if err != nil {
		return fmt.Errorf("cannot open file %s: %w", w.path, err)
	}

	currentValue := string(bytes.TrimSpace(data))
	changed := false
	w.mu.Lock()
	if currentValue != w.value {
		w.value = currentValue
		changed = true
	}
	w.mu.Unlock()

	if notify && changed && w.onChange != nil {
		w.onChange(w)
	}

	return nil
}

// Writes new value into the watchd file. Does not update new value into watched file, requires calling of Read() method.
func (w *WatchedFile) Write(value string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	file, err := os.OpenFile(w.path, os.O_WRONLY, 0)
	if err != nil {
		return fmt.Errorf("cannot open file %s: %w", w.path, err)
	}
	defer file.Close()

	cleanValue := strings.TrimSpace(value)

	_, err = file.WriteString(cleanValue)
	if err != nil {
		return fmt.Errorf("cannot write '%s' to %s: %w", cleanValue, w.path, err)
	}

	return nil
}
