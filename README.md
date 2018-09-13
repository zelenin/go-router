# go-router

Deadly simple router

## Usage

```go
package main

import (
    "net/http"
    "log"
    "github.com/zelenin/go-router"
)

func main() {
    rtr := router.New()

    rtr.Add("/login", []string{"GET", "POST"}, loginHandler)

    rtr.Get("/user/:username", func(res http.ResponseWriter, req *http.Request) {
        params := router.Params(req)
        res.Write([]byte("Username: " + params.Get("username")))
    })

    postRtr := router.New()

    // /posts/{:id:[\d]+}
    postRtr.Get(`/{:id:[\d]+}`, postHandler)

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

```

## Notes

* WIP. Library API can be changed in the future

## Author

[Aleksandr Zelenin](https://github.com/zelenin/), e-mail: [aleksandr@zelenin.me](mailto:aleksandr@zelenin.me)
