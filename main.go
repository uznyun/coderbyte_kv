package main

import (
	"context"
	"net/http"
	"sync"
)

type KVStorage interface {
	Get(ctx context.Context, key string) (interface{}, error)
	Put(ctx context.Context, key string, val interface{}) error
	Delete(ctx context.Context, key string) error
}

// -------

type storage struct {
	mu    sync.RWMutex
	store map[string]interface{}
}

func NewStorage() *storage {
	return &storage{
		store: make(map[string]interface{}),
	}
}

func (s *storage) Get(ctx context.Context, key string) (interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.store[key], nil
}

func (s *storage) Put(ctx context.Context, key string, val interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.store[key] = val

	return nil
}

func (s *storage) Delete(ctx context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.store, key)

	return nil
}

// -------

type counter struct {
	kv KVStorage
}

func NewCounter(kv KVStorage) *counter {
	return &counter{kv}
}

func (c *counter) increaseHandler(res http.ResponseWriter, req *http.Request) {
	const key = "key"

	ctx := req.Context()
	raw, _ := c.kv.Get(ctx, key)
	if raw == nil {
		_ = c.kv.Put(ctx, key, 0)
		return
	}

	val := raw.(int)
	val++
	_ = c.kv.Put(ctx, key, val)
}

// -------

func main() {
	kv := NewStorage()
	counter := NewCounter(kv)

	http.HandleFunc("/inc", counter.increaseHandler)
	_ = http.ListenAndServe(":3000", nil)
}
