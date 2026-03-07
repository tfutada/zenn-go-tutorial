# Go Maps — Basics, Memory Leak, and Weak Pointers

## Basics

```go
// Create with make
m := make(map[string]int)
m["key"] = 42

// Create with literal
m := map[string]int{"key": 42}
```

- Accessing a missing key returns the **zero value** (no panic)
- Use the **comma-ok idiom** (`val, ok := m[key]`) to check existence
- `delete(m, key)` removes an entry (no-op if key doesn't exist)
- Iteration order is **not guaranteed**
- A **nil map** can be read but panics on write

## Map Memory Leak (Map Churn)

**Map churn** = repeatedly inserting and deleting entries over time (e.g., a session cache that grows during peak hours and shrinks after).

Go maps **never shrink their internal bucket array**. After a map grows to N buckets, it keeps that memory even if most keys are deleted. This applies to both the Swiss table (Go 1.24+) and the old bucket implementation equally.

### Demo result

```
initial:                          HeapInuse =   576 KB
after churn (100 entries left):   HeapInuse = 295184 KB   <-- 288 MB for 100 entries!
after replace + copy:             HeapInuse =   568 KB
```

### Fix: rebuild the map

```go
newMap := make(map[K]V, len(oldMap))
for k, v := range oldMap {
    newMap[k] = v
}
oldMap = newMap
```

### Cross-language comparison

| | Go | Rust | Java | Node.js (V8) |
|---|---|---|---|---|
| Auto-shrink on delete | No | No | No | Yes |
| Manual shrink | No (rebuild) | `shrink_to_fit()` | No (rebuild) | Not needed |
| When it shrinks | Never | On demand | Never | When deleted slots > threshold |

- **Rust** uses the same Swiss table algorithm as Go 1.24+, but exposes `shrink_to_fit()`
- **Node.js (V8)** auto-compacts when enough "hole" entries accumulate, at the cost of occasional latency spikes
- **Go and Java** require manual map replacement

## Weak Pointers (Go 1.24+)

`weak.Pointer[T]` is Go's equivalent of Java's `WeakReference<T>`. The GC can collect the target when no strong references remain.

```go
entry := &CacheEntry{Data: "hello"}
wp := weak.Make(entry)

// Later...
if v := wp.Value(); v != nil {
    // still alive
} else {
    // collected by GC
}
```

### Comparison with Java

| Java | Go 1.24+ |
|---|---|
| `WeakReference<T>` | `weak.Pointer[T]` |
| `.get()` returns `null` | `.Value()` returns `nil` |
| `WeakHashMap` (built-in) | No equivalent (build it yourself) |
| `SoftReference` | No equivalent |
| `ReferenceQueue` | No equivalent (must poll `.Value()`) |

You must **manually clean up** stale entries from a weak pointer cache:

```go
for k, wp := range cache {
    if wp.Value() == nil {
        delete(cache, k)
    }
}
```

## Run

```bash
go run src/hash_map/main.go
```
