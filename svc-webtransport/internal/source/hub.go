// CLASSIFICATION: UNCLASSIFIED
// hub is the non-blocking fan-out primitive behind Source. It maintains the set
// of per-session subscriber channels and delivers each record to every one of
// them without blocking on a slow consumer.
package source

import (
	"context"
	"sync"
)

// hub fans records out to a dynamic set of subscriber channels.
type hub struct {
	mu      sync.RWMutex
	subs    map[chan []byte]struct{}
	bufSize int
}

// newHub creates a hub whose subscriber channels each buffer bufSize records.
func newHub(bufSize int) *hub {
	if bufSize <= 0 {
		bufSize = 1
	}
	return &hub{
		subs:    make(map[chan []byte]struct{}),
		bufSize: bufSize,
	}
}

// subscribe registers a new subscriber channel and returns it. The channel is
// unregistered and closed when ctx is cancelled.
func (h *hub) subscribe(ctx context.Context) <-chan []byte {
	ch := make(chan []byte, h.bufSize)

	h.mu.Lock()
	h.subs[ch] = struct{}{}
	h.mu.Unlock()

	go func() {
		<-ctx.Done()
		h.remove(ch)
	}()
	return ch
}

// remove unregisters and closes a subscriber channel if still present. Closing
// happens under the write lock, mutually exclusive with broadcast's read lock,
// so a send can never race with a close.
func (h *hub) remove(ch chan []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.subs[ch]; ok {
		delete(h.subs, ch)
		close(ch)
	}
}

// broadcast delivers b to every subscriber, dropping the record for any
// subscriber whose buffer is full. It returns the delivered and dropped counts.
func (h *hub) broadcast(b []byte) (delivered, dropped int) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for ch := range h.subs {
		select {
		case ch <- b:
			delivered++
		default:
			dropped++
		}
	}
	return delivered, dropped
}

// subscriberCount returns the number of active subscribers.
func (h *hub) subscriberCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.subs)
}

// closeAll unregisters and closes every subscriber channel.
func (h *hub) closeAll() {
	h.mu.Lock()
	defer h.mu.Unlock()
	for ch := range h.subs {
		delete(h.subs, ch)
		close(ch)
	}
}
