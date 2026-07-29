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
	latency := flag.Duration("latency", 0, "simulated latency")
	errorRate := flag.Float64("error-rate", 0, "fraction of requests returning 500")
	useH2C := flag.Bool("h2c", false, "enable HTTP/2 cleartext (h2c)")
	flag.Parse()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("proto: %s", r.Proto)
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

	srv := &http.Server{
		Addr:    ":" + *port,
		Handler: mux,
	}
	if *useH2C {
		srv.Protocols = new(http.Protocols)
		srv.Protocols.SetHTTP1(true)
		srv.Protocols.SetUnencryptedHTTP2(true)
		log.Printf("mockserver listening on :%s (h2c enabled)\n", *port)
	} else {
		log.Printf("mockserver listening on :%s\n", *port)
	}
	log.Fatal(srv.ListenAndServe())
}
