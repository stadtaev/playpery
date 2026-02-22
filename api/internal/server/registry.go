package server

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/playperu/cityquiz/internal/database"
)

type Registry struct {
	dir    string
	mu     sync.RWMutex
	stores map[string]*DocStore
}

func NewRegistry(dir string) *Registry {
	return &Registry{
		dir:    dir,
		stores: make(map[string]*DocStore),
	}
}

func (r *Registry) Get(ctx context.Context, slug string) (*DocStore, error) {
	r.mu.RLock()
	s, ok := r.stores[slug]
	r.mu.RUnlock()
	if ok {
		return s, nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Double-check after acquiring write lock.
	if s, ok := r.stores[slug]; ok {
		return s, nil
	}

	s, err := r.open(ctx, slug)
	if err != nil {
		return nil, err
	}
	r.stores[slug] = s
	return s, nil
}

func (r *Registry) Create(ctx context.Context, slug string) (*DocStore, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if s, ok := r.stores[slug]; ok {
		return s, nil
	}

	s, err := r.open(ctx, slug)
	if err != nil {
		return nil, err
	}
	r.stores[slug] = s
	return s, nil
}

func (r *Registry) open(ctx context.Context, slug string) (*DocStore, error) {
	dbPath := filepath.Join(r.dir, slug+".db")
	db, err := database.Open(ctx, dbPath)
	if err != nil {
		return nil, fmt.Errorf("opening client db %q: %w", slug, err)
	}
	store, err := NewDocStore(ctx, db)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("initializing client store %q: %w", slug, err)
	}
	return store, nil
}

func (r *Registry) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for slug, s := range r.stores {
		s.db.Close()
		delete(r.stores, slug)
	}
	return nil
}
