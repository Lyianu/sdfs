// main.go generates binary for master(metadata) servers
package main

import (
	"log"
	"net/http"

	"github.com/Lyianu/sdfs/router"
)

func main() {
	r := router.NewRouter()
	log.Printf("Starting Master, Listening on %s", ":8080")
	http.ListenAndServe(":8080", r)
}
