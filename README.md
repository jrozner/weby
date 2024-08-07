# Weby

Weby is a very simple server side web framework that intends to simply wrap the Go http library and leverage the routing functionality introduced in Go 1.22. It provides some basic middleware that would be commonly utilized and intends to stay completely compatible with the standard library rather than introducing framework specific interfaces and types like many other frameworks have had to do to introduce routers, contexts, and other features that are now part of the standard library.

## Use

```go
package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/jrozner/weby"
	"github.com/jrozner/weby/middleware"
	"github.com/jrozner/weby/rlog"
)

func main() {
	var handler slog.Handler = slog.NewTextHandler(os.Stdout, nil)
	handler = rlog.RequestIDHandler{handler}
	logger := slog.New(handler)
	mux := weby.NewServeMux()
	mux.Use(middleware.RequestID)
	mux.Use(middleware.WrapResponse)
	mux.Use(middleware.Logger(logger))
	mux.HandleFunc("/", root)

	log.Fatal(http.ListenAndServe(":8080", mux))
}

func root(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
```

## License

Weby is licensed under [MIT](LICENSE) though you probably shouldn't use it for your own projects.