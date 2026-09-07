package syspower

import (
	"fmt"
	"os"
	"slices"
	"strings"
)

type SysFsVal struct {
	valueFile *WatchedFile
	choices   []string
}

// Creates new SysFsVal object with static predefined slice of choices
func NewStaticFsVal(valueFile *WatchedFile, choices []string) *SysFsVal {
	return &SysFsVal{valueFile: valueFile, choices: choices}
}

// Creates new SysFsVal object with from file loaded slice of choices
func NewDynamicFsVal(valueFile *WatchedFile, choicesPath string) (*SysFsVal, error) {
	data, err := os.ReadFile(choicesPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read choices file %s: %w", choicesPath, err)
	}

	choices := strings.Fields(string(data))

	return &SysFsVal{valueFile: valueFile, choices: choices}, nil
}

// Returns last loaded value of SysFs
func (s *SysFsVal) Value() string {
	return s.valueFile.Value()
}

// Returns copy of possible choices
func (s *SysFsVal) Choices() []string {
	cp := make([]string, len(s.choices))
	copy(cp, s.choices)
	return cp
}

// Sets new value into SysFs. Only allowed choices will be set.
func (s *SysFsVal) Set(value string) error {
	cleanedValue := strings.TrimSpace(value)

	if slices.Contains(s.choices, cleanedValue) {
		return s.valueFile.Write(cleanedValue)
	}
	return fmt.Errorf("value '%s' is not listed in possible choices %v", cleanedValue, s.choices)
}

// Invoked read methond on watched file to update stored value.
func (s *SysFsVal) UpdateValue() error {
	return s.valueFile.Read()
}
