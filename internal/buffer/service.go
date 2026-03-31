package buffer

import (
	"sync"
	"time"
)

type BufItem any

type Service[T any] struct {
	buffer      chan T
	ready       chan struct{}
	maxSize     int
	flushPeriod time.Duration
	mu          sync.RWMutex
	closed      bool
}

const defaultMaxBufferSize = 50

// NewService creates a new buffer service
func NewService[T any](maxSize int, flushPeriod time.Duration) *Service[T] {
	if maxSize <= 0 {
		maxSize = defaultMaxBufferSize
	}

	service := &Service[T]{
		buffer:      make(chan T, maxSize),
		ready:       make(chan struct{}, 1), // Buffered to prevent blocking
		maxSize:     maxSize,
		flushPeriod: flushPeriod,
	}

	// Start periodic flush
	go service.startPeriodicFlush()

	return service
}

func (s *Service[T]) Push(msg T) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return false
	}

	select {
	case s.buffer <- msg:
		// Check if buffer is nearly full
		if len(s.buffer) >= s.maxSize-1 {
			s.signalReady()
		}
		return true
	default:
		// Buffer is full
		return false
	}
}

func (s *Service[T]) Pop() []T {
	s.mu.RLock()
	closed := s.closed
	s.mu.RUnlock()

	// Even if closed, we might want to drain what's left
	// But the original code returns nil if closed.
	// Let's stick to returning nil if closed, or maybe drain if closed?
	// Usually if closed, we want to drain the last bits.
	if closed && len(s.buffer) == 0 {
		return nil
	}

	capacity := len(s.buffer)
	if capacity == 0 {
		capacity = 0
	}
	result := make([]T, 0, capacity)

	// Drain the buffer
	for {
		select {
		case item, ok := <-s.buffer:
			if !ok {
				return result
			}
			result = append(result, item)
		default:
			return result
		}
	}
}

func (s *Service[T]) IsEmpty() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.buffer) == 0
}

func (s *Service[T]) Size() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.buffer)
}

func (s *Service[T]) Ready() <-chan struct{} {
	return s.ready
}

func (s *Service[T]) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.closed {
		s.closed = true
		// We don't close the channels to avoid panics in Push/signalReady.
		// Instead, we rely on s.closed check.
	}
}

func (s *Service[T]) signalReady() {
	select {
	case s.ready <- struct{}{}:
	default:
	}
}

func (s *Service[T]) startPeriodicFlush() {
	ticker := time.NewTicker(s.flushPeriod)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.RLock()
		closed := s.closed
		s.mu.RUnlock()

		if closed {
			return
		}

		if len(s.buffer) > 0 {
			s.signalReady()
		}
	}
}
