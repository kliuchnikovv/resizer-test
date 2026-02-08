package main

import (
	"context"
	"log"
	"net/http"

	lru "github.com/hashicorp/golang-lru"
)

const (
	proto    = "http://"
	hostport = "localhost:8080"
	success  = "success"
	failure  = "failure"
)

func main() {
	cache, err := lru.New(1024)
	if err != nil {
		log.Panicf("Faild to create cache: %v", err)
	}

	svc := &service{
		cache:           cache,
		numberOfWorkers: 10,

		retrieveAttempts: 5,
		retrieveTimeoutS: 10,
	}

	ctx := context.Background()
	if err := svc.startWorkers(ctx); err != nil {
		log.Panicf("Faild to start workers: %v", err)
	}
	defer svc.stopWorkers()

	mux := http.NewServeMux()
	mux.Handle("/v1/resize", svc.resizeHandler())
	mux.Handle("/v1/image/", svc.getImageHandler())
	address := hostport

	log.Print("Listening on ", hostport)
	// When running on docker mac, can't listen only on localhost
	panic(http.ListenAndServe(address, mux))
}
