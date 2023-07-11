package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/pprof"
)

func startHttpProfile(port int) {
	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/pprof/", pprof.Index)
		mux.HandleFunc("/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("/pprof/profile", pprof.Profile)
		mux.HandleFunc("/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("/pprof/trace", pprof.Trace)

		err := http.ListenAndServe(fmt.Sprintf(":%d", port), mux)
		if err != nil {
			log.Printf("perf error:%+v", err)
		}
	}()
}
