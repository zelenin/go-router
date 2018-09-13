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

func WithNotFoundHandler(notFoundHandler http.Handler) Option {
    return func(router *Router) {
        router.notFoundHandler = notFoundHandler
    }
}

func New(options ...Option) *Router {
    rtr := &Router{
        routes: map[string]*route{},
        cache:  newPipeCache(),
    }

    for _, option := range options {
        option(rtr)
    }

    if rtr.notFoundHandler == nil {
        rtr.notFoundHandler = http.NotFoundHandler()
    }

    return rtr
}

type route struct {
    pattern   string
    rePattern *regexp.Regexp
    methods   []string
    handler   http.Handler
}

type middleware struct {
    pattern   string
    rePattern *regexp.Regexp
    methods   []string
    handler   func(http.Handler) http.Handler
}

type pipe struct {
    route       *route
    middlewares []*middleware
}

type Router struct {
    routes          map[string]*route
    middlewares     []*middleware
    notFoundHandler http.Handler
    cache           *pipeCache
}

func (router *Router) ServeHTTP(res http.ResponseWriter, req *http.Request) {
    params := &RequestParams{}
    req = req.WithContext(context.WithValue(req.Context(), "params", params))

    pipe := router.match(req)

    handler := router.notFoundHandler
    if pipe.route != nil {
        handler = pipe.route.handler
    }

    chainMiddleware(pipe.middlewares...)(handler).ServeHTTP(res, req)
}

func (router *Router) Get(pattern string, handler http.Handler) {
    router.route(pattern, []string{http.MethodGet}, handler)
}

func (router *Router) Head(pattern string, handler http.Handler) {
    router.route(pattern, []string{http.MethodHead}, handler)
}

func (router *Router) Post(pattern string, handler http.Handler) {
    router.route(pattern, []string{http.MethodPost}, handler)
}

func (router *Router) Put(pattern string, handler http.Handler) {
    router.route(pattern, []string{http.MethodPut}, handler)
}

func (router *Router) Patch(pattern string, handler http.Handler) {
    router.route(pattern, []string{http.MethodPatch}, handler)
}

func (router *Router) Delete(pattern string, handler http.Handler) {
    router.route(pattern, []string{http.MethodDelete}, handler)
}

func (router *Router) Connect(pattern string, handler http.Handler) {
    router.route(pattern, []string{http.MethodConnect}, handler)
}

func (router *Router) Options(pattern string, handler http.Handler) {
    router.route(pattern, []string{http.MethodOptions}, handler)
}

func (router *Router) Trace(pattern string, handler http.Handler) {
    router.route(pattern, []string{http.MethodTrace}, handler)
}

func (router *Router) All(pattern string, handler http.Handler) {
    router.route(pattern, []string{MethodAll}, handler)
}

func (router *Router) SubRoute(pattern string, subRouter *Router) {
    for _, route := range subRouter.routes {
        router.route(pattern+route.pattern, route.methods, route.handler)
    }
    for _, middleware := range subRouter.middlewares {
        router.pipe(pattern+middleware.pattern, middleware.methods, middleware.handler)
    }
}

func (router *Router) Route(pattern string, methods []string, handler http.Handler) {
    router.route(pattern, methods, handler)
}

func (router *Router) route(pattern string, methods []string, handler http.Handler) {
    if pattern == "" {
        log.Panicf("http: invalid pattern '%s'", pattern)
    }
    if pattern[0] != '/' {
        log.Panicf("pattern must begin with '/' in '%s'", pattern)
    }
    if handler == nil {
        log.Panicf("http: nil handler")
    }
    _, ok := router.routes[pattern]
    if ok {
        log.Panicf("http: multiple registrations for %s", pattern)
    }

    router.routes[pattern] = &route{
        pattern:   pattern,
        rePattern: regexp.MustCompile(normalizePattern(pattern)),
        methods:   methods,
        handler:   handler,
    }
}

func (router *Router) Pipe(pattern string, handler func(http.Handler) http.Handler) {
    router.pipe(pattern, []string{MethodAll}, handler)
}

func (router *Router) pipe(pattern string, methods []string, handler func(http.Handler) http.Handler) {
    if pattern == "" {
        log.Panicf("http: invalid pattern '%s'", pattern)
    }
    if pattern[0] != '/' {
        log.Panicf("pattern must begin with '/' in '%s'", pattern)
    }
    if handler == nil {
        log.Panicf("http: nil handler")
    }

    router.middlewares = append(router.middlewares, &middleware{
        pattern:   pattern,
        rePattern: regexp.MustCompile(normalizePattern(pattern)),
        methods:   methods,
        handler:   handler,
    })
}

func (router *Router) match(req *http.Request) *pipe {
    pipe, err := router.cache.Get(req)
    if err != nil {
        pipe = match(req, router.routes, router.middlewares)
        router.cache.Set(req, pipe)
    }

    if pipe.route != nil {
        injectParameters(req, pipe)
    }

    return pipe
}

func match(req *http.Request, routes map[string]*route, middlewares []*middleware) *pipe {
    maxPatternLen := 0
    var matchedRoute *route

    for _, route := range routes {
        if check(req, route.methods, route.rePattern) {
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

    pipe := &pipe{
        route:       matchedRoute,
        middlewares: []*middleware{},
    }

    for _, middleware := range middlewares {
        if check(req, middleware.methods, middleware.rePattern) {
            pipe.middlewares = append(pipe.middlewares, middleware)
        }
    }

    return pipe
}

func check(req *http.Request, methods []string, rePattern *regexp.Regexp) bool {
    if !matchMethod(req.Method, methods) {
        return false
    }

    if !rePattern.MatchString(req.URL.Path) {
        return false
    }

    return true
}

func matchMethod(method string, methods []string) bool {
    for _, routeMethod := range methods {
        if method == routeMethod || routeMethod == MethodAll {
            return true
        }
    }

    return false
}

func injectParameters(req *http.Request, pipe *pipe) {
    params := Params(req)

    matches := pipe.route.rePattern.FindStringSubmatch(req.URL.Path)

    names := pipe.route.rePattern.SubexpNames()
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

func chainMiddleware(middlewares ...*middleware) func(http.Handler) http.Handler {
    return func(final http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            last := final
            for i := len(middlewares) - 1; i >= 0; i-- {
                last = middlewares[i].handler(last)
            }
            last.ServeHTTP(w, r)
        })
    }
}
