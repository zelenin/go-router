package router

import (
    "errors"
    "net/http"
    "sync"
)

var errNotFound = errors.New("not found")

// @todo invalidation
type pipeCache struct {
    mu    sync.Mutex
    pipes map[string]map[string]*pipe
}

func (cache *pipeCache) Get(req *http.Request) (*pipe, error) {
    cache.mu.Lock()
    defer cache.mu.Unlock()

    _, ok := cache.pipes[req.Method]
    if !ok {
        return nil, errNotFound
    }

    _, ok = cache.pipes[req.Method][req.URL.String()]
    if !ok {
        return nil, errNotFound
    }

    return cache.pipes[req.Method][req.URL.String()], nil
}

func (cache *pipeCache) Set(req *http.Request, p *pipe) {
    cache.mu.Lock()
    defer cache.mu.Unlock()

    _, ok := cache.pipes[req.Method]
    if !ok {
        cache.pipes[req.Method] = map[string]*pipe{}
    }

    cache.pipes[req.Method][req.URL.String()] = p
}

func newPipeCache() *pipeCache {
    return &pipeCache{
        pipes: map[string]map[string]*pipe{},
    }
}
