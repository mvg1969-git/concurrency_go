package engine

import (
	"context"
	"sync"
)

type Engine struct {
	sync.RWMutex
	kv map[string]string
}

func NewEngine() *Engine {
	return &Engine{
		kv: make(map[string]string),
	}
}

func (e *Engine) Set(ctx context.Context, key, value string) {
	e.Lock()
	defer e.Unlock()
	e.kv[key] = value
}

func (e *Engine) Get(ctx context.Context, key string) (string, bool) {
	e.RLock()
	defer e.RUnlock()
	val, found := e.kv[key]
	return val, found
}

func (e *Engine) Del(ctx context.Context, key string) {
	e.Lock()
	defer e.Unlock()
	delete(e.kv, key)
}
