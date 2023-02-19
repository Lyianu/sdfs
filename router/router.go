package router

import (
	"fmt"
	"net/http"
)

// Router communicates with master and client
type Router struct {
}

func NewRouter() *Router {
	return new(Router)
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	fmt.Fprint(w, "Hello")
}
