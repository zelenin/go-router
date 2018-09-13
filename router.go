package router

import (
    "bytes"
    "context"
    lexer "github.com/zelenin/go-router/pattern"
    "log"
    "net/http"
    "regexp"
    "strings"
)

const MethodAll = "*"

type Option func(*Router)

func WithNotFoundHandler(notFoundHandler http.HandlerFunc) Option {
    return func(router *Router) {
        router.notFoundHandler = notFoundHandler
    }
}

func New(options ...Option) *Router {
    rtr := &Router{
        routes: map[string]*route{},
        cache:  newRouteCache(),
    }

    for _, option := range options {
        option(rtr)
    }

    if rtr.notFoundHandler == nil {
        rtr.notFoundHandler = http.NotFound
    }

    return rtr
}

type route struct {
    pattern   string
    rePattern *regexp.Regexp
    methods   []string
    handler   http.HandlerFunc
}

type Router struct {
    routes          map[string]*route
    notFoundHandler http.HandlerFunc
    cache           *routeCache
}

func (router *Router) ServeHTTP(res http.ResponseWriter, req *http.Request) {
    params := &RequestParams{}
    req = req.WithContext(context.WithValue(req.Context(), "params", params))

    route := router.match(req)
    if route != nil {
        route.handler(res, req)
    } else {
        router.notFoundHandler.ServeHTTP(res, req)
    }
}

func (router *Router) Get(pattern string, handler http.HandlerFunc) {
    router.add(pattern, []string{http.MethodGet}, handler)
}

func (router *Router) Head(pattern string, handler http.HandlerFunc) {
    router.add(pattern, []string{http.MethodHead}, handler)
}

func (router *Router) Post(pattern string, handler http.HandlerFunc) {
    router.add(pattern, []string{http.MethodPost}, handler)
}

func (router *Router) Put(pattern string, handler http.HandlerFunc) {
    router.add(pattern, []string{http.MethodPut}, handler)
}

func (router *Router) Patch(pattern string, handler http.HandlerFunc) {
    router.add(pattern, []string{http.MethodPatch}, handler)
}

func (router *Router) Delete(pattern string, handler http.HandlerFunc) {
    router.add(pattern, []string{http.MethodDelete}, handler)
}

func (router *Router) Connect(pattern string, handler http.HandlerFunc) {
    router.add(pattern, []string{http.MethodConnect}, handler)
}

func (router *Router) Options(pattern string, handler http.HandlerFunc) {
    router.add(pattern, []string{http.MethodOptions}, handler)
}

func (router *Router) Trace(pattern string, handler http.HandlerFunc) {
    router.add(pattern, []string{http.MethodTrace}, handler)
}

func (router *Router) All(pattern string, handler http.HandlerFunc) {
    router.add(pattern, []string{MethodAll}, handler)
}

func (router *Router) SubRoute(pattern string, subRouter *Router) {
    for _, route := range subRouter.routes {
        router.add(pattern+route.pattern, route.methods, route.handler)
    }
}

func (router *Router) Add(pattern string, methods []string, handler http.HandlerFunc) {
    router.add(pattern, methods, handler)
}

func (router *Router) add(pattern string, methods []string, handler http.HandlerFunc) {
    if pattern == "" {
        log.Panicf("http: invalid pattern '%s'", pattern)
    }
    if pattern[0] != '/' {
        log.Panicf("pattern must begin with '/' in '%s'", pattern)
    }
    if handler == nil {
        log.Panicf("http: nil handler")
    }
    if _, ok := router.routes[pattern]; ok {
        log.Panicf("http: multiple registrations for %s", pattern)
    }

    router.routes[pattern] = &route{
        pattern:   pattern,
        rePattern: regexp.MustCompile(normalizePattern(pattern)),
        methods:   methods,
        handler:   handler,
    }
}

func (router *Router) match(req *http.Request) *route {
    route, err := router.cache.Get(req)
    if err != nil {
        route = match(req, router.routes)
        router.cache.Set(req, route)
    }

    if route != nil {
        injectParameters(req, route)
    }

    return route
}

func match(req *http.Request, routes map[string]*route) *route {
    maxPatternLen := 0
    var matchedRoute *route

    for _, route := range routes {
        if matchRoute(req, route) {
            patternLen := len(route.pattern)

            if patternLen < maxPatternLen {
                continue
            }

            if patternLen > maxPatternLen || (patternLen == maxPatternLen && route.pattern > matchedRoute.pattern) {
                matchedRoute = route
                maxPatternLen = patternLen
            }
        }
    }

    return matchedRoute
}

func matchRoute(req *http.Request, route *route) bool {
    if !matchMethod(req.Method, route.methods) {
        return false
    }

    if !route.rePattern.MatchString(req.URL.Path) {
        return false
    }

    return true
}

func matchMethod(method string, routeMethods []string) bool {
    for _, routeMethod := range routeMethods {
        if method == routeMethod || routeMethod == MethodAll {
            return true
        }
    }

    return false
}

func injectParameters(req *http.Request, route *route) {
    params := Params(req)

    matches := route.rePattern.FindStringSubmatch(req.URL.Path)

    names := route.rePattern.SubexpNames()
    for i, match := range matches {
        if i != 0 {
            params.Set(names[i], match)
        }
    }
}

func normalizePattern(pattern string) string {
    tokenizer := lexer.Lex(bytes.NewBufferString(pattern))

    patternBuf := bytes.NewBufferString("^")

    for {
        token := tokenizer.NextToken()
        if token.Type == lexer.TOKEN_EOF {
            break
        }

        switch token.Type {
        case lexer.TOKEN_LEFT_CURLY_BRACKET:
            handleCurlyParam(tokenizer, patternBuf)

        case lexer.TOKEN_COLON:
            handleColonParam(tokenizer, patternBuf)

        default:
            patternBuf.Write(token.Value)
        }
    }

    return strings.Replace(patternBuf.String(), `//`, `/`, -1)
}

const tagNamePattern = `[a-zA-Z][\w]*`

func handleCurlyParam(tokenizer lexer.Tokenizer, patternBuf *bytes.Buffer) {
    token := tokenizer.NextToken()

    tagName := token.String()
    pattern := tagNamePattern

    token = tokenizer.NextToken()

    if token.Type == lexer.TOKEN_PARAM_SEPARATOR {
        token := tokenizer.NextToken()

        pattern = token.String()

        tokenizer.NextToken()
    }

    patternBuf.WriteString(`(?P<` + tagName + `>` + pattern + `)`)
}

func handleColonParam(tokenizer lexer.Tokenizer, patternBuf *bytes.Buffer) {
    token := tokenizer.NextToken()

    tagName := token.String()
    pattern := tagNamePattern

    patternBuf.WriteString(`(?P<` + tagName + `>` + pattern + `)`)
}

type RequestParams map[string]string

func (params *RequestParams) Set(key string, value string) {
    (*params)[key] = value
}

func (params *RequestParams) Get(key string) string {
    return (*params)[key]
}

func Params(req *http.Request) *RequestParams {
    return req.Context().Value("params").(*RequestParams)
}
