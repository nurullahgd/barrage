package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"
)

func main() {
	port := flag.String("port", "8090", "port to listen on")
	latency := flag.Duration("latency", 0, "simulated latency (e.g. 10ms)")
	errorRate := flag.Float64("error-rate", 0, "fraction of requests that return 500 (e.g. 0.05 for 5%%)")
	flag.Parse()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if *latency > 0 {
			time.Sleep(*latency)
		}
		if *errorRate > 0 && rand.Float64() < *errorRate {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"status":"ok"}`)
	})

	log.Printf("mockserver listening on :%s (latency=%s error-rate=%.0f%%)\n", *port, *latency, *errorRate*100)
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}
