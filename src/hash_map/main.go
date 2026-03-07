package main

import (
	"fmt"
	"runtime"
	"weak"
)

func main() {
	// Creating a map using make
	mymap := make(map[string]int)
	mymap["hp"] = 1000
	mymap["dell"] = 800
	mymap["lenovo"] = 600
	fmt.Println("make:", mymap)
	fmt.Println("hp:", mymap["hp"])

	// Creating a map using a map literal
	prices := map[string]int{
		"hp":     1000,
		"dell":   800,
		"lenovo": 600,
	}
	fmt.Println("literal:", prices)

	// Accessing a non-existent key returns the zero value
	fmt.Println("apple:", prices["apple"]) // 0

	// Check if a key exists using the comma-ok idiom
	val, ok := prices["dell"]
	fmt.Println("dell:", val, "exists:", ok)

	val, ok = prices["apple"]
	fmt.Println("apple:", val, "exists:", ok)

	// Update an existing key
	prices["hp"] = 1200
	fmt.Println("updated hp:", prices["hp"])

	// Delete a key
	delete(prices, "lenovo")
	fmt.Println("after delete:", prices)

	// Iterate over a map (order is not guaranteed)
	fmt.Println("--- iterate ---")
	for k, v := range prices {
		fmt.Printf("  %s: $%d\n", k, v)
	}

	// Length of a map
	fmt.Println("len:", len(prices))

	// A nil map can be read but not written to
	var nilmap map[string]int
	fmt.Println("nil map read:", nilmap["key"]) // 0
	fmt.Println("nil map len:", len(nilmap))     // 0
	// nilmap["key"] = 1 // this would panic

	// --- Map memory leak demo ---
	// Maps never shrink their internal bucket array.
	// After churn (grow then partial delete), memory stays at peak size.
	// This affects both Swiss table (Go 1.24+) and old bucket maps equally.
	fmt.Println("\n--- map memory leak demo ---")
	demoMapMemoryLeak()

	// --- Weak pointer cache ---
	// Go 1.24+ provides weak.Pointer, similar to Java's WeakReference.
	// GC can collect values when no strong references remain.
	fmt.Println("\n--- weak pointer cache ---")
	demoWeakCache()
}

func demoMapMemoryLeak() {
	printAlloc := func(label string) {
		var m runtime.MemStats
		runtime.GC()
		runtime.ReadMemStats(&m)
		fmt.Printf("  %-35s HeapInuse = %d KB\n", label, m.HeapInuse/1024)
	}

	// Simulate a cache that grows to 1M entries, then shrinks to 100.
	// The map retains its peak bucket allocation even with only 100 entries.
	cache := make(map[int][128]byte)
	printAlloc("initial:")

	for i := 0; i < 1_000_000; i++ {
		cache[i] = [128]byte{}
	}
	// Shrink back to 100 entries
	for i := 100; i < 1_000_000; i++ {
		delete(cache, i)
	}
	printAlloc("after churn (100 entries left):")
	// ~288 MB retained for just 100 entries!

	// Fix: replace the map and copy surviving entries
	newCache := make(map[int][128]byte, len(cache))
	for k, v := range cache {
		newCache[k] = v
	}
	cache = newCache
	printAlloc("after replace + copy:")
	_ = cache
}

type CacheEntry struct {
	Data string
}

func demoWeakCache() {
	// A cache using weak pointers: entries are collected when no one else holds them
	cache := make(map[string]weak.Pointer[CacheEntry])

	// Create entries with strong references
	entry1 := &CacheEntry{Data: "important data"}
	entry2 := &CacheEntry{Data: "temporary data"}

	cache["kept"] = weak.Make(entry1)
	cache["temp"] = weak.Make(entry2)

	fmt.Println("  before GC:")
	printWeakCache(cache)

	// Drop the strong reference to entry2; entry1 is still held
	entry2 = nil
	runtime.GC()

	fmt.Println("  after GC (dropped temp):")
	printWeakCache(cache)

	// Clean up stale entries — this is your responsibility with weak pointers
	for k, wp := range cache {
		if wp.Value() == nil {
			delete(cache, k)
		}
	}
	fmt.Printf("  cache size after cleanup: %d\n", len(cache))

	// KeepAlive ensures entry1 is not collected before this point
	runtime.KeepAlive(entry1)
}

func printWeakCache(cache map[string]weak.Pointer[CacheEntry]) {
	for k, wp := range cache {
		if v := wp.Value(); v != nil {
			fmt.Printf("    %s -> %s\n", k, v.Data)
		} else {
			fmt.Printf("    %s -> (collected by GC)\n", k)
		}
	}
}
