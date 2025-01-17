# go-router

Deadly simple router

## Usage

```go
package main

import (
	"github.com/zelenin/go-router"
	"log"
	"net/http"
)

func main() {
	rtr := router.New()

	rtr.Pipe(logger)

	rtr.HandleFunc("GET /ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})

	rtr.Group("/api/", func(subRouter *router.Router) {
		subRouter.Group("/posts/", func(subRouter *router.Router) {
			subRouter.Handle(`GET /{id}`, postHandler{})
		})
	})

	server := &http.Server{
		Addr:    ":8080",
		Handler: rtr,
	}

	err := server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}

type postHandler struct{}

func (handler postHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	id := req.PathValue("id")
	res.Write([]byte("id: " + id))
}

func logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[logger] Connection from %s", r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}
```

## Notes

* WIP. Library API can be changed in the future

## Author

[Aleksandr Zelenin](https://github.com/zelenin/), e-mail: [aleksandr@zelenin.me](mailto:aleksandr@zelenin.me)
