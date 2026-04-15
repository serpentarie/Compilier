package interpreter

import "fmt"

type runtimeSlot struct {
	value       Object
	initialized bool
	used        bool
}

type Environment struct {
	values    map[string]*runtimeSlot
	enclosing *Environment
}

func NewEnvironment(enclosing *Environment) *Environment {
	return &Environment{
		values:    make(map[string]*runtimeSlot),
		enclosing: enclosing,
	}
}

func (e *Environment) Define(name string) {
	e.values[name] = &runtimeSlot{initialized: false, used: false}
}

func (e *Environment) DefineInitialized(name string, value Object) {
	e.values[name] = &runtimeSlot{value: value, initialized: true, used: false}
}

func (e *Environment) Assign(name string, value Object) error {
	if slot, ok := e.values[name]; ok {
		slot.value = value
		slot.initialized = true
		return nil
	}

	if e.enclosing != nil {
		return e.enclosing.Assign(name, value)
	}

	return fmt.Errorf("runtime error: variable '%s' is not declared", name)
}

func (e *Environment) Get(name string) (Object, error) {
	if slot, ok := e.values[name]; ok {
		slot.used = true
		if !slot.initialized {
			return Object{}, fmt.Errorf("runtime error: variable '%s' is used before initialization", name)
		}
		return slot.value, nil
	}

	if e.enclosing != nil {
		return e.enclosing.Get(name)
	}

	return Object{}, fmt.Errorf("runtime error: variable '%s' is not declared", name)
}

func (e *Environment) CollectUnusedWarnings() []error {
	warnings := make([]error, 0)
	for name, slot := range e.values {
		if !slot.used {
			warnings = append(warnings, fmt.Errorf("[WARNING] runtime: variable '%s' was declared but never used", name))
		}
	}
	return warnings
}
