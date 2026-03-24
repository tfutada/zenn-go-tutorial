// SSE (Server-Sent Events) example demonstrating lightweight real-time
// server-to-client push using only the Go standard library.
//
// Run:
//   go run src/sse/main.go
//
// Then open http://localhost:8080 in a browser.
// Multiple browser tabs = multiple concurrent SSE connections.
package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// broker manages SSE client connections and broadcasts events.
type broker struct {
	mu      sync.RWMutex
	clients map[chan string]struct{}
}

func newBroker() *broker {
	return &broker{clients: make(map[chan string]struct{})}
}

func (b *broker) subscribe() chan string {
	ch := make(chan string, 16) // buffered to avoid blocking broadcast
	b.mu.Lock()
	b.clients[ch] = struct{}{}
	b.mu.Unlock()
	log.Printf("client connected (total: %d)", b.count())
	return ch
}

func (b *broker) unsubscribe(ch chan string) {
	b.mu.Lock()
	delete(b.clients, ch)
	close(ch)
	b.mu.Unlock()
	log.Printf("client disconnected (total: %d)", b.count())
}

func (b *broker) broadcast(data string) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for ch := range b.clients {
		select {
		case ch <- data:
		default:
			// slow client, drop event to avoid blocking
		}
	}
}

func (b *broker) count() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.clients)
}

// sseHandler streams events to the client over a long-lived HTTP connection.
func sseHandler(b *broker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// SSE required headers
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("X-Accel-Buffering", "no") // disable nginx buffering

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming not supported", http.StatusInternalServerError)
			return
		}

		ch := b.subscribe()
		defer b.unsubscribe(ch)

		// send initial connection event
		fmt.Fprintf(w, "event: connected\ndata: {\"status\":\"ok\"}\n\n")
		flusher.Flush()

		for {
			select {
			case <-r.Context().Done():
				// client disconnected
				return
			case msg := <-ch:
				fmt.Fprintf(w, "data: %s\n\n", msg)
				flusher.Flush()
			}
		}
	}
}

// index page with embedded EventSource client
func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, indexHTML)
}

const indexHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<title>Go SSE Demo</title>
<style>
  body { font-family: monospace; max-width: 700px; margin: 2rem auto; background: #1a1a2e; color: #e0e0e0; }
  h1 { color: #00d2ff; }
  #status { padding: 4px 8px; border-radius: 4px; font-size: 0.9rem; }
  .connected { background: #0f5132; color: #75b798; }
  .disconnected { background: #842029; color: #ea868f; }
  #events { list-style: none; padding: 0; }
  #events li { padding: 6px 10px; margin: 4px 0; background: #16213e; border-left: 3px solid #00d2ff; }
  #events li .time { color: #888; margin-right: 8px; }
</style>
</head>
<body>
  <h1>SSE Demo</h1>
  <p>Status: <span id="status" class="disconnected">connecting...</span></p>
  <p>Events received: <span id="count">0</span></p>
  <ul id="events"></ul>
<script>
  const status = document.getElementById('status');
  const countEl = document.getElementById('count');
  const events = document.getElementById('events');
  let count = 0;

  const source = new EventSource('/events');

  source.addEventListener('connected', () => {
    status.textContent = 'connected';
    status.className = 'connected';
  });

  source.onmessage = (e) => {
    count++;
    countEl.textContent = count;
    const li = document.createElement('li');
    const now = new Date().toLocaleTimeString();
    li.innerHTML = '<span class="time">' + now + '</span>' + e.data;
    events.prepend(li);
    // keep last 100 events in DOM
    while (events.children.length > 100) events.lastChild.remove();
  };

  source.onerror = () => {
    status.textContent = 'disconnected (reconnecting...)';
    status.className = 'disconnected';
  };
</script>
</body>
</html>`

func main() {
	b := newBroker()

	// background goroutine: broadcast a timestamp every second
	go func() {
		id := 0
		for {
			time.Sleep(1 * time.Second)
			id++
			msg := fmt.Sprintf(`{"id":%d,"time":"%s","clients":%d}`,
				id, time.Now().Format(time.RFC3339), b.count())
			b.broadcast(msg)
		}
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("/", indexHandler)
	mux.HandleFunc("/events", sseHandler(b))

	addr := ":8080"
	log.Printf("SSE server listening on http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
