package syspower

import (
	"fmt"
	"os"
	"slices"
	"time"
)

type ControlSpec struct {
	path          string
	choicesPath   string
	staticChoices []string
}

type Registry struct {
	specs       map[string]ControlSpec
	specsNames  []string
	cache       map[string]*SysFsVal
	cachedNames []string
}

func NewRegistry() *Registry {
	r := &Registry{specs: make(map[string]ControlSpec), cache: make(map[string]*SysFsVal)}

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

	if f, exists := r.cache[name]; exists {
		return f, nil
	}

	watched := NewWatchedFile(s.path, onChange)

	if s.choicesPath != func() string { return "" }() {
		if _, err := os.Stat(s.choicesPath); err == nil {
			if fsVal, err := NewDynamicFsVal(watched, s.choicesPath); err != nil {
				return nil, err
			} else {
				r.cache[name] = fsVal
				r.cachedNames = append(r.cachedNames, name)
				return fsVal, nil
			}
		}
	}

	fsVal := NewStaticFsVal(watched, s.staticChoices)
	r.cachedNames = append(r.cachedNames, name)
	slices.Sort(r.cachedNames)
	r.cache[name] = fsVal
	return fsVal, nil
}

// List all known control specification names.
func (r *Registry) List() []string {
	var result []string
	for _, name := range r.specsNames {
		result = append(result, name)
	}
	return result
}

// List all successfully cached control names.
func (r *Registry) ListCached() []string {
	var result []string
	for _, name := range r.cachedNames {
		result = append(result, name)
	}
	return result
}

// Watch periodically for changes.
func (r *Registry) WatchChanges() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		<-ticker.C
		for _, name := range r.ListCached() {
			if ctrl, err := r.Get(name, nil); err == nil {
				ctrl.UpdateValue()
			}
		}
	}
}
