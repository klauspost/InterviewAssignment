package traffic

import "sync"

// syncErr is a synchronized error value that can safely be
// accessed from multiple goroutines.
// Only the first error will be kept.
type syncErr struct {
	err error
	mu  sync.RWMutex
}

// Err will return the first error that has been set.
func (e *syncErr) Err() error {
	e.mu.RLock()
	err := e.err
	e.mu.RUnlock()
	return err
}

// Set will set the error to a given error.
// If nil is passed, it is ignored.
// If an error value has already been set, it will
// be ignored.
func (e *syncErr) Set(err error) {
	if err == nil {
		return
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.err != nil {
		return
	}
	e.err = err
}
