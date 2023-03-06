package main

import (
	"log"
	"net/http"

	"github.com/Lyianu/sdfs/driver"
	"github.com/Lyianu/sdfs/router"
)

func main() {
	r := router.NewRouter()
	log.Printf("Starting Node, Listening on %s", ":8080")
	http.ListenAndServe(":8080", r)
	_ = driver.File{}

}
