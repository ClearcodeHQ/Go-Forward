// +build debug

package main

import (
	"net/http"
	_ "net/http/pprof"
)

func debug() {
	go func() {
		http.ListenAndServe("localhost:6060", nil)
	}()
}
