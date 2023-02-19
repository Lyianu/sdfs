package main

import (
	"net/http"

	"github.com/Lyianu/sdfs/router"
)

func main() {
	r := router.NewRouter()
	http.ListenAndServe(":8080", r)
}
