package main

import (
	"fmt"
	"github.com/gosuri/uilive"
	"net/http"
	"runtime"
	"sync/atomic"
	"time"
)

func hello(w http.ResponseWriter, req *http.Request) {
	ch <- true
	atomic.AddInt32(&count, 1)
	time.Sleep(time.Second * 10)
	_, _ = fmt.Fprintf(w, "hello\n")
	atomic.AddInt32(&count, -1)
	ch <- true
}

var count int32
var ch chan bool

func main() {
	ch = make(chan bool)

	go func() {
		var m runtime.MemStats
		var writer = uilive.New()
		writer.Start()
		defer writer.Stop()

		for {
			<-ch
			_, _ = fmt.Fprintf(writer, "Current connections count: %d\n", atomic.LoadInt32(&count))
			runtime.ReadMemStats(&m)
			_, _ = fmt.Fprintf(writer, "Alloc = %v MiB\n", m.Alloc/1024/1024)
			_, _ = fmt.Fprintf(writer, "TotalAlloc = %v MiB\n", m.TotalAlloc/1024/1024)
			_, _ = fmt.Fprintf(writer, "Sys = %v MiB\n", m.Sys/1024/1024)
			_, _ = fmt.Fprintf(writer, "NumGC = %v\n", m.NumGC)
		}
	}()

	http.HandleFunc("/", hello)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		return
	}
}
