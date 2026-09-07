package syspower

import (
	"fmt"
	"os"
)

type ControlSpec struct {
	path          string
	choicesPath   string
	staticChoices []string
}

type Registry struct {
	specs map[string]ControlSpec
}

func NewRegistry() *Registry {
	r := &Registry{specs: make(map[string]ControlSpec)}

	r.specs["cpuboost"] = ControlSpec{
		path:          "/sys/devices/system/cpu/cpufreq/boost",
		staticChoices: []string{"0", "1"},
	}

	r.specs["platform_profile"] = ControlSpec{
		path:        "/sys/firmware/acpi/platform_profile",
		choicesPath: "/sys/firmware/acpi/platform_profile_choices",
	}

	return r
}

func (r *Registry) Get(name string, onChange OnChange) (*SysFsVal, error) {
	s, exists := r.specs[name]
	if !exists {
		return nil, fmt.Errorf("unknown control '%s'", name)
	}

	if _, err := os.Stat(s.path); os.IsNotExist(err) {
		return nil, fmt.Errorf("control '%s' is not supported on this hardware (missing %s)", name, s.path)
	}

	watched := NewWatchedFile(s.path, onChange)

	if s.choicesPath != func() string { return "" }() {
		if _, err := os.Stat(s.choicesPath); err == nil {
			return NewDynamicFsVal(watched, s.choicesPath)
		}
	}

	return NewStaticFsVal(watched, s.staticChoices), nil
}

func (r *Registry) List() []string {
	var result []string
	for name, _ := range r.specs {
		result = append(result, name)
	}
	return result
}
