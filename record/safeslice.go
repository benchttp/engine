package record

import "sync"

// SafeSlice holds a slice of Record that is safe for concurrent use.
type SafeSlice struct {
	mu   sync.Mutex
	data []Record
}

// NewSafeSlice returns a SafeSlice initizalized with a capacity
// set to the given size.
func NewSafeSlice(size int) SafeSlice {
	return SafeSlice{
		mu:   sync.Mutex{},
		data: make([]Record, 0, size),
	}
}

// Append safely appends a Record to SafeSlice in a concurrent context.
func (s *SafeSlice) Append(rec Record) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data = append(s.data, rec)
}

// Slice returns a Go slice containing the Records.
func (s *SafeSlice) Slice() []Record {
	return s.data
}
