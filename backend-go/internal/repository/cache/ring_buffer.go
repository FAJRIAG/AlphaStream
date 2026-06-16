// Package cache provides an in-memory, thread-safe circular buffer for OHLCV data.
// The RingBuffer eliminates the need for frequent DB reads in the hot path of
// the Quantitative Engine, minimizing both latency and memory allocation.
package cache

import (
	"sync"

	"github.com/alphastream/backend-go/internal/domain/entity"
)

// DefaultRingCapacity is the number of OHLCV candles kept in memory per symbol.
// 500 candles covers MA-50 warm-up (50) + RSI-14 warm-up (14) with plenty of headroom.
const DefaultRingCapacity = 500

// RingBuffer is a fixed-capacity circular buffer for entity.OHLCV values.
// It is safe for concurrent use via a read-write mutex:
//   - Multiple goroutines may Read (Slice) simultaneously.
//   - Only one goroutine may Write (Push) at a time.
type RingBuffer struct {
	mu       sync.RWMutex
	buf      []entity.OHLCV
	capacity int
	head     int // Index of the oldest element
	tail     int // Index where the next element will be written
	count    int // Number of valid elements currently in the buffer
}

// NewRingBuffer allocates a RingBuffer with the given capacity.
// Panics if capacity is less than 1 to catch configuration errors early.
func NewRingBuffer(capacity int) *RingBuffer {
	if capacity < 1 {
		panic("cache.NewRingBuffer: capacity must be >= 1")
	}
	return &RingBuffer{
		buf:      make([]entity.OHLCV, capacity),
		capacity: capacity,
	}
}

// Push appends a new OHLCV candle to the buffer.
// If the buffer is full, the oldest entry is overwritten (circular behavior).
// Time complexity: O(1).
func (rb *RingBuffer) Push(candle entity.OHLCV) {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	rb.buf[rb.tail] = candle
	rb.tail = (rb.tail + 1) % rb.capacity

	if rb.count == rb.capacity {
		// Buffer is full — advance head to discard the oldest entry.
		rb.head = (rb.head + 1) % rb.capacity
	} else {
		rb.count++
	}
}

// Slice returns all valid elements in the buffer in chronological order
// (oldest first). The returned slice is a fresh copy — callers may safely
// iterate without holding the lock.
// Time complexity: O(n).
func (rb *RingBuffer) Slice() []entity.OHLCV {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	if rb.count == 0 {
		return nil
	}

	result := make([]entity.OHLCV, rb.count)
	for i := 0; i < rb.count; i++ {
		result[i] = rb.buf[(rb.head+i)%rb.capacity]
	}
	return result
}

// Last returns the most recently pushed candle.
// Returns the zero-value OHLCV and false if the buffer is empty.
func (rb *RingBuffer) Last() (entity.OHLCV, bool) {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	if rb.count == 0 {
		return entity.OHLCV{}, false
	}

	// tail points to the *next write* slot, so last written is tail-1.
	lastIdx := (rb.tail - 1 + rb.capacity) % rb.capacity
	return rb.buf[lastIdx], true
}

// Count returns the number of elements currently in the buffer.
func (rb *RingBuffer) Count() int {
	rb.mu.RLock()
	defer rb.mu.RUnlock()
	return rb.count
}

// ─── RingBufferStore ──────────────────────────────────────────────────────────

// RingBufferStore manages one RingBuffer per stock symbol.
// It is the in-memory cache layer between the Simulator and the Quantitative Engine.
type RingBufferStore struct {
	mu       sync.RWMutex
	buffers  map[string]*RingBuffer
	capacity int
}

// NewRingBufferStore creates a store with per-symbol buffers of the given capacity.
func NewRingBufferStore(capacity int) *RingBufferStore {
	return &RingBufferStore{
		buffers:  make(map[string]*RingBuffer),
		capacity: capacity,
	}
}

// Push appends a candle to the buffer for the given symbol.
// If no buffer exists for the symbol, one is created lazily.
func (s *RingBufferStore) Push(symbol string, candle entity.OHLCV) {
	s.mu.Lock()
	buf, ok := s.buffers[symbol]
	if !ok {
		buf = NewRingBuffer(s.capacity)
		s.buffers[symbol] = buf
	}
	s.mu.Unlock()

	// Push outside the store lock to minimize contention between symbols.
	buf.Push(candle)
}

// GetSlice returns a chronological copy of all candles for a symbol.
// Returns nil if the symbol has no buffer yet.
func (s *RingBufferStore) GetSlice(symbol string) []entity.OHLCV {
	s.mu.RLock()
	buf, ok := s.buffers[symbol]
	s.mu.RUnlock()

	if !ok {
		return nil
	}
	return buf.Slice()
}

// GetLast returns the most recent candle for a symbol.
func (s *RingBufferStore) GetLast(symbol string) (entity.OHLCV, bool) {
	s.mu.RLock()
	buf, ok := s.buffers[symbol]
	s.mu.RUnlock()

	if !ok {
		return entity.OHLCV{}, false
	}
	return buf.Last()
}

// Count returns the number of candles buffered for a symbol.
func (s *RingBufferStore) Count(symbol string) int {
	s.mu.RLock()
	buf, ok := s.buffers[symbol]
	s.mu.RUnlock()

	if !ok {
		return 0
	}
	return buf.Count()
}

// Clear deletes the ring buffer for a given symbol.
func (s *RingBufferStore) Clear(symbol string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.buffers, symbol)
}
