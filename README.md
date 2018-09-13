# go-router

Deadly simple router

## Usage

```go
package main

import (
    "github.com/zelenin/go-router"
    "log"
    "net/http"
    "strings"
)

func main() {
    rtr := router.New()

    rtr.Route("/login", []string{"GET", "POST"}, http.HandlerFunc(loginHandler))

    rtr.Get("/user/:username", http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
        params := router.Params(req)
        res.Write([]byte("Username: " + params.Get("username")))
    }))

    rtr.Pipe("/", logger)

    postRtr := router.New()

    // /posts/{id:[\d]+}
    postRtr.Get(`/{id:[\d]+}`, postHandler{})

    // /posts/comments
    postRtr.Get("/comments", commentsHandler)

    // /posts/comments/add
    postRtr.Post(`/comments/add`, addCommentsHandler)

    rtr.SubRoute("/posts", postRtr)

    server := &http.Server{
        Addr:    ":8080",
        Handler: rtr,
    }

    err := server.ListenAndServe()
    if err != nil {
        log.Fatal(err)
    }
}

func loginHandler(res http.ResponseWriter, req *http.Request) {
    res.Write([]byte("Login page"))
}

type postHandler struct{}

func (handler postHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
    res.Write([]byte("Post page"))
}

func logger(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        log.Printf("Connection from %s", r.RemoteAddr)
        next.ServeHTTP(w, r)
    })
}

```

## Notes

* WIP. Library API can be changed in the future

## Author

[Aleksandr Zelenin](https://github.com/zelenin/), e-mail: [aleksandr@zelenin.me](mailto:aleksandr@zelenin.me)
