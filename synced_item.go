package prago

import "sync"

type syncedItem[T any] struct {
	mu  sync.RWMutex
	val T
}

// Set updates the time safely
func (st *syncedItem[T]) Set(t T) {
	st.mu.Lock()
	defer st.mu.Unlock()
	st.val = t
}

// Get retrieves the current time safely
func (st *syncedItem[T]) Get() T {
	st.mu.RLock()
	defer st.mu.RUnlock()
	return st.val
}
