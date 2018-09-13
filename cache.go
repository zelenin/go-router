package router

import (
    "errors"
    "net/http"
    "sync"
)

var errNotFound = errors.New("not found")

type routeCache struct {
    mu     sync.Mutex
    routes map[string]map[string]*route
}

func (cache *routeCache) Get(req *http.Request) (*route, error) {
    cache.mu.Lock()
    defer cache.mu.Unlock()

    if _, ok := cache.routes[req.Method]; !ok {
        return nil, errNotFound
    }

    if _, ok := cache.routes[req.Method][req.URL.String()]; !ok {
        return nil, errNotFound
    }

    return cache.routes[req.Method][req.URL.String()], nil
}

func (cache *routeCache) Set(req *http.Request, r *route) {
    cache.mu.Lock()
    defer cache.mu.Unlock()

    if _, ok := cache.routes[req.Method]; !ok {
        cache.routes[req.Method] = map[string]*route{}
    }

    cache.routes[req.Method][req.URL.String()] = r
}

func newRouteCache() *routeCache {
    return &routeCache{
        routes: map[string]map[string]*route{},
    }
}
