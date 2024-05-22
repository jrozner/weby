# Weby

Weby is a very simple server side web framework that intends to simply wrap the Go http library and leverage the routing functionality introduced in Go 1.22. It provides some basic middleware that would be commonly utilized and intends to stay completely compatible with the standard library rather than introducing framework specific interfaces and types like many other frameworks have had to do to introduce routers, contexts, and other features that are now part of the standard library.

## Use

```go
package main

import (
	"log"
	"net/http"

	"github.com/jrozner/weby"
	"github.com/jrozner/weby/middleware"
)

func main() {
	server := weby.NewServer()
	server.Use(middleware.WrapResponse)
	server.Use(middleware.Logger(log.Default()))
	server.HandleFunc("/", root)
	
	log.Fatal(http.ListenAndServe(":8080", server))
}

func root(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)	
}
```

## License

Weby is licensed under [MIT](LICENSE) though you probably shouldn't use it for your own projects.