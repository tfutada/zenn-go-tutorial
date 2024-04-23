package main

import (
	"encoding/base64"
	"fmt"
	"github.com/gosuri/uilive"
	"golang.org/x/crypto/sha3"
	"math/rand"
	"net/http"
	"runtime"
	"sync/atomic"
	"time"
)

func hello(w http.ResponseWriter, req *http.Request) {
	ch <- true // dump stats
	atomic.AddInt32(&count, 1)
	time.Sleep(time.Second * 3)
	//_, _ = fmt.Fprintf(w, "hello\n")
	workerChann <- "work"
	ret := <-workerChann
	_, _ = fmt.Fprintf(w, "hash: %s\n", ret)
	atomic.AddInt32(&count, -1)
	ch <- true // dump stats
}

// channels are thread-safe
var count int32
var ch chan bool
var workerChann chan string

func main() {
	ch = make(chan bool)
	workerChann = make(chan string)

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

	// spawn a worker goroutine for getHashOfRandomString, which receive a task via channel
	go func() {
		for {
			task := <-workerChann
			workerChann <- getHashOfRandomString(task)
		}
	}()

	http.HandleFunc("/", hello)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		return
	}
}

// mimic cpu heavy work
func getHashOfRandomString(task string) string {
	var d = make([]byte, 32)
	rand.Read(d)
	for i := 0; i < 1000000; i++ {
		result := sha3.Sum512(d)
		d = result[:]
	}
	return base64.StdEncoding.EncodeToString(d)
}
