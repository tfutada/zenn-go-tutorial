package main

import (
	"fmt"
	"runtime"
	"runtime/debug"
)

func main() {
	leak := leakExample()
	fmt.Println("Leaky SubSlice:", string(leak))
	runtime.GC()
	debug.FreeOSMemory()
	fmt.Printf("Memory after leak: %d MB\n", currentAllocKB()/1024)

	fixed := leakExampleFixed()
	fmt.Println("Fixed SubSlice:", string(fixed))
	runtime.GC()
	debug.FreeOSMemory()
	fmt.Printf("Memory after fixed: %d MB\n", currentAllocKB()/1024)
}

func leakExample() []byte {
	// 100 MB to make effect visible
	data := make([]byte, 100*1024*1024)
	copy(data, "This is a long byte slice that we will use to demonstrate a slice leak.")
	return data[:10]
}

func leakExampleFixed() []byte {
	data := make([]byte, 100*1024*1024)
	copy(data, "This is a long byte slice that we will use to demonstrate a slice leak.")

	sub := data[:10]
	result := make([]byte, len(sub))
	copy(result, sub)
	return result
}

func currentAllocKB() uint64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.Alloc / 1024
}
