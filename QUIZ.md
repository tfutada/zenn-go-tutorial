# Go Tutorial Quiz

Test your understanding of the Go concepts demonstrated in this repository.

---

## Question 1: Goroutines and Channels
Based on `src/goroutine/main.go`:

```go
func main() {
    ch := make(chan string)
    go processTask(ch)
    fmt.Println("waiting...")
    result := <-ch
    fmt.Println("Received result:", result)
}

func processTask(ch chan<- string) {
    time.Sleep(2 * time.Second)
    ch <- "Hello"
}
```

**What will happen if you remove the line `result := <-ch`?**

A) The program will print "waiting..." and exit immediately
B) The program will wait 2 seconds then print the result
C) The program will panic
D) The program will deadlock

<details>
<summary>Answer</summary>

**A) The program will print "waiting..." and exit immediately**

Without `result := <-ch`, the main goroutine doesn't wait for the channel to receive a value. The main function completes and exits before the goroutine can send "Hello" to the channel.
</details>

---

## Question 2: Race Conditions
Based on `src/race1/main.go`:

```go
type Pair struct {
    X int
    Y int
}

arr := []Pair{{X: 0, Y: 0}, {X: 1, Y: 1}}
var p Pair

// writer
go func() {
    for i := 0; ; i++ {
        p = arr[i%2]
    }
}()

// reader
for {
    read := p
    switch read.X + read.Y {
    case 0, 2:
        // Expected cases
    default:
        return fmt.Sprintf("struct corruption detected: %+v", read)
    }
}
```

**Why can this code detect "struct corruption" even though we only assign `{X: 0, Y: 0}` or `{X: 1, Y: 1}`?**

A) The compiler optimizes the code incorrectly
B) Race condition: the reader might read X and Y at different times during assignment
C) The array index calculation is wrong
D) Slices are not thread-safe in Go

<details>
<summary>Answer</summary>

**B) Race condition: the reader might read X and Y at different times during assignment**

Struct assignment is not atomic. The writer goroutine might be halfway through assigning the struct when the reader reads it, resulting in values like `{X: 0, Y: 1}` or `{X: 1, Y: 0}`, which sum to 1 instead of 0 or 2.
</details>

---

## Question 3: Context and Cancellation
Based on `src/context/main.go`:

```go
func main() {
    ctx := context.Background()
    resultsCh := make(chan *WorkResult)

    childCtx, cancel := context.WithCancel(ctx)
    defer cancel()

    go doWork(childCtx, resultsCh, 50)

    select {
    case <-time.After(3 * time.Second):
        fmt.Println("Timeout occurred before work was completed.")
        cancel()
    case result := <-resultsCh:
        fmt.Printf("Work completed: %#v\n", result)
    }
}
```

**What happens if the Fibonacci calculation (n=50) takes longer than 3 seconds?**

A) The program waits until Fibonacci completes
B) The program prints "Timeout occurred" and the goroutine continues in the background
C) The program prints "Timeout occurred" and the goroutine detects cancellation
D) The program panics

<details>
<summary>Answer</summary>

**C) The program prints "Timeout occurred" and the goroutine detects cancellation**

After 3 seconds, the timeout case triggers, `cancel()` is called, and the `doWork` function's `select` statement will detect `ctx.Done()` and print "Work was canceled."
</details>

---

## Question 4: Interface Nil Check
Based on `src/interface/main.go`:

```go
var x interface{}
var y *int = nil
x = y

if x != nil {
    fmt.Println("x != nil")
} else {
    fmt.Println("x == nil")
}

v := reflect.ValueOf(x).IsNil()
fmt.Println(v)
```

**What is the output?**

A) `x == nil` and `false`
B) `x != nil` and `true`
C) `x == nil` and `true`
D) `x != nil` and `false`

<details>
<summary>Answer</summary>

**B) `x != nil` and `true`**

An interface value is nil only if both its type and value are nil. When you assign `y` (a `*int` with nil value) to `x`, the interface now has a type (`*int`) and a value (nil). So `x != nil` is true (the interface itself is not nil), but `reflect.ValueOf(x).IsNil()` returns true (the underlying value is nil).
</details>

---

## Question 5: Slice Memory Management
Based on `src/slice1/main.go`:

```go
func readFileDetails(name string) ([]byte, error) {
    data, err := os.ReadFile(name)
    if err != nil {
        return nil, err
    }
    //return data[5:10], nil
    return bytes.Clone(data[5:10]), nil
}
```

**Why is `bytes.Clone()` used instead of just returning `data[5:10]`?**

A) To make the code run faster
B) To prevent the entire file contents from being kept in memory
C) `bytes.Clone()` is required for slice operations
D) To create a deep copy of the slice elements

<details>
<summary>Answer</summary>

**B) To prevent the entire file contents from being kept in memory**

When you slice `data[5:10]`, the returned slice still references the underlying array of the full file. Using `bytes.Clone()` creates a new backing array with only the needed bytes, allowing the original large buffer to be garbage collected.
</details>

---

## Question 6: Functional Options Pattern
Based on `src/functional_options/main.go`:

```go
type Server struct {
    IP   string
    Port int
}

type OptionsServerFunc func(c *Server) error

func NewServer(opts ...OptionsServerFunc) (*Server, error) {
    s := &Server{}
    for _, opt := range opts {
        if err := opt(s); err != nil {
            return nil, err
        }
    }
    return s, nil
}

func WithIP(ip string) OptionsServerFunc {
    return func(s *Server) error {
        s.IP = ip
        return nil
    }
}
```

**What is the main benefit of the functional options pattern?**

A) It makes the code run faster
B) It allows optional configuration without multiple constructor functions
C) It's required for struct initialization in Go
D) It prevents race conditions

<details>
<summary>Answer</summary>

**B) It allows optional configuration without multiple constructor functions**

The functional options pattern provides a clean way to handle optional parameters and configuration. You can have many optional settings without creating multiple constructor functions or dealing with config structs.
</details>

---

## Question 7: Empty Struct Memory
Based on `src/empty_struct/main.go`:

```go
type Set map[string]struct{}

mySet := make(Set)
mySet["apple"] = struct{}{}
mySet["banana"] = struct{}{}
```

**Why use `struct{}` as the map value type instead of `bool`?**

A) `struct{}` is faster than `bool`
B) `struct{}` takes zero bytes of memory
C) `struct{}` is a Go requirement for sets
D) `struct{}` prevents race conditions

<details>
<summary>Answer</summary>

**B) `struct{}` takes zero bytes of memory**

`struct{}` is an empty struct that occupies zero bytes of memory. When implementing a set, you only care about the keys, not the values. Using `struct{}` instead of `bool` saves 1 byte per map entry.
</details>

---

## Question 8: Producer-Consumer Pattern
Based on `src/producer_consumer/main.go`:

```go
func main() {
    dataChan := make(chan int, 10)

    go consumeData(dataChan)

    for i := 0; i < 10; i++ {
        fmt.Printf("Producing %d\n", i)
        dataChan <- i
        time.Sleep(100 * time.Millisecond)
    }

    close(dataChan)
}

func consumeData(in <-chan int) {
    for num := range in {
        fmt.Printf("Consumed %d\n", num)
        time.Sleep(1 * time.Second)
    }
}
```

**What happens when main() finishes?**

A) The program waits for the consumer to finish all items
B) The program exits immediately, potentially before the consumer finishes
C) The program deadlocks
D) The program panics because the channel is closed

<details>
<summary>Answer</summary>

**B) The program exits immediately, potentially before the consumer finishes**

The consumer runs in a goroutine. When `main()` finishes (after producing all items and closing the channel), the program exits even if the consumer hasn't finished processing. The consumer processes much slower (1 second per item) than the producer (100ms per item).
</details>

---

## Bonus Question: Channel Directionality

In the goroutine example, `processTask` has the signature:
```go
func processTask(ch chan<- string)
```

**What does `chan<-` mean?**

A) The channel can only receive values
B) The channel can only send values
C) The channel is bidirectional
D) The channel is buffered

<details>
<summary>Answer</summary>

**B) The channel can only send values**

`chan<- string` is a send-only channel. The function can only send to this channel, not receive from it. This provides compile-time safety by restricting operations on the channel.
</details>

---

## Scoring Guide

- 8-9 correct: Expert level! You understand Go concurrency and patterns deeply.
- 6-7 correct: Advanced level. You have a strong grasp of Go concepts.
- 4-5 correct: Intermediate level. Keep practicing with the examples.
- 0-3 correct: Beginner level. Review the source code and run the examples!
