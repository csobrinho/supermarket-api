package supermarket

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/google/logger"
	"golang.org/x/exp/maps"
)

type Factory interface {
	// CreateSupermarket creates a new supermarket instance.
	Create(ctx context.Context, name string, opts ...Option) (Supermarket, error)
	// RegisterSupermarket registers a new supermarket creator function.
	Register(name string, creator Creator)
	// Available returns a list of registered supermarkets.
	Available() []string
}

// Creator is a function that creates a supermarket instance.
type Creator func(ctx context.Context, cfg *Config) (Supermarket, error)

type factory struct {
	mu        sync.RWMutex
	factories map[string]Creator
}

// NewFactory creates a new factory instance.
func NewFactory() Factory {
	return &factory{
		factories: make(map[string]Creator),
	}
}

// Register registers a new supermarket creator function.
func (f *factory) Register(name string, creator Creator) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.factories[name] = creator
	logger.Infof("supermarket: registered %q provider", name)
}

// Create creates a new supermarket instance.
func (f *factory) Create(ctx context.Context, name string, opts ...Option) (Supermarket, error) {
	f.mu.RLock()
	creator, exists := f.factories[name]
	f.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("supermarket: %q is not registered", name)
	}
	cfg := &Config{Timeout: 30 * time.Second}
	for _, opt := range opts {
		opt(cfg)
	}
	return creator(ctx, cfg)
}

// AvailableSupermarkets returns a list of registered supermarkets.
func (f *factory) Available() []string {
	f.mu.RLock()
	defer f.mu.RUnlock()

	names := maps.Keys(f.factories)
	sort.Strings(names)
	return names
}
