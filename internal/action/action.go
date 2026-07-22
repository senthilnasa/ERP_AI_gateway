package action

import (
	"fmt"
	"sync"
)

type Action interface {
	Name() string
	Description() string
}

type DefaultAction struct {
	name        string
	description string
}

func New(name string, description string) *DefaultAction {
	return &DefaultAction{
		name:        name,
		description: description,
	}
}

func (a *DefaultAction) Name() string {
	return a.name
}

func (a *DefaultAction) Description() string {
	return a.description
}

type Registry struct {
	mu      sync.RWMutex
	actions map[string]Action
}

func NewRegistry() *Registry {
	r := &Registry{
		actions: make(map[string]Action),
	}

	// Pre-register standard actions specified in requirement spec
	r.Register(New("rewrite", "Rewrites text into a given tone and style"))
	r.Register(New("summarize", "Summarizes text into key points or executive summary"))
	r.Register(New("translate", "Translates text into target language"))
	r.Register(New("improve", "Improves clarity, vocabulary, and grammar"))
	r.Register(New("expand", "Elaborates and adds details to text"))
	r.Register(New("shorten", "Condenses text while maintaining key points"))
	r.Register(New("proofread", "Fixes typos, spelling, and grammar errors"))

	return r
}

func (r *Registry) Register(a Action) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.actions[a.Name()] = a
}

func (r *Registry) Get(name string) (Action, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	a, exists := r.actions[name]
	if !exists {
		return nil, fmt.Errorf("action '%s' is not registered", name)
	}
	return a, nil
}

func (r *Registry) ListNames() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.actions))
	for name := range r.actions {
		names = append(names, name)
	}
	return names
}
