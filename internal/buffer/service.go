package buffer

import (
	"sync"
	"time"
)

type BufItem interface{}

type Service[T BufItem] struct {
	buffer      chan T
	ready       chan struct{}
	maxSize     int
	flushPeriod time.Duration
	mu          sync.RWMutex
	closed      bool
}

const defaultMaxBufferSize = 50

// NewService creates a new buffer service
func NewService[T BufItem](maxSize int, flushPeriod time.Duration) *Service[T] {
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
	defer s.mu.RUnlock()

	if s.closed {
		return nil
	}

	result := make([]T, 0)

	// Drain the buffer
	for {
		select {
		case item := <-s.buffer:
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
		close(s.buffer)
		close(s.ready)
	}
}

func (s *Service[T]) signalReady() {
	select {
	case s.ready <- struct{}{}:
	default:
		// Channel is full, signal already sent
	}
}

func (s *Service[T]) startPeriodicFlush() {
	ticker := time.NewTicker(s.flushPeriod)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.RLock()
		closed := s.closed
		empty := len(s.buffer) == 0
		s.mu.RUnlock()

		if closed {
			return
		}

		if !empty {
			s.signalReady()
		}
	}
}
