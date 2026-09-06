package actions

import (
	"fmt"
	"sync"
)

// Registry maintains all registered business actions.
type Registry interface {
	Register(action Action) error
	GetAction(name string) (Action, error)
	ListActions() []Action
}

type defaultRegistry struct {
	mu      sync.RWMutex
	actions map[string]Action
}

// NewRegistry creates a new Action Registry.
func NewRegistry() Registry {
	return &defaultRegistry{
		actions: make(map[string]Action),
	}
}

func (r *defaultRegistry) Register(action Action) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.actions[action.Name()]; exists {
		return fmt.Errorf("action %s is already registered", action.Name())
	}

	r.actions[action.Name()] = action
	return nil
}

func (r *defaultRegistry) GetAction(name string) (Action, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	action, exists := r.actions[name]
	if !exists {
		return nil, fmt.Errorf("action %s not found", name)
	}

	return action, nil
}

func (r *defaultRegistry) ListActions() []Action {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var list []Action
	for _, action := range r.actions {
		list = append(list, action)
	}

	return list
}
