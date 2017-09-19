package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"numbers"
)

func main() {
	listenAddr := flag.String("http.addr", ":8080", "http listen address")
	responseTimeout := flag.Int("timeout.response", 480, "server response timeout (in ms)")
	getTimeout := flag.Int("timeout.geturl", 450, "timeout for URL get calls (in ms)")
	numGoRoutines := flag.Int("goroutine.count", 20, "concurrency factor")

	flag.Parse()

	ng := &numbers.NumbersGetter{}
	ng.ResponseTimeout = time.Duration(*responseTimeout) * time.Millisecond
	ng.GetTimeout = time.Duration(*getTimeout) * time.Millisecond
	ng.NumGoRoutines = *numGoRoutines

	http.Handle("/numbers", ng)
	log.Fatal(http.ListenAndServe(*listenAddr, nil))
}
