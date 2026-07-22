package profile

import (
	"fmt"
	"sync"

	"github.com/senthilnasa/ERP_AI_gateway/internal/models"
	"github.com/senthilnasa/ERP_AI_gateway/internal/prompt"
)

type Profile interface {
	Name() string
	Validate(req *models.WriteRequest) error
	BuildPrompt(req *models.WriteRequest, engine *prompt.PromptEngine) (string, error)
	ParseResponse(raw string) (string, error)
}

type Registry struct {
	mu       sync.RWMutex
	profiles map[string]Profile
}

func NewRegistry() *Registry {
	return &Registry{
		profiles: make(map[string]Profile),
	}
}

func (r *Registry) Register(p Profile) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.profiles[p.Name()] = p
}

func (r *Registry) Get(name string) (Profile, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, exists := r.profiles[name]
	if !exists {
		return nil, fmt.Errorf("profile '%s' is not registered", name)
	}
	return p, nil
}

func (r *Registry) ListNames() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.profiles))
	for name := range r.profiles {
		names = append(names, name)
	}
	return names
}
