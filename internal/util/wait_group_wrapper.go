package util

import "sync"

// WaitGroupWrapper is a extension for sync.WaitGroup.
type WaitGroupWrapper struct {
	sync.WaitGroup
}

// Wrap is a helper for basic case to sync.WaitGroup.
func (w *WaitGroupWrapper) Wrap(cb func()) {
	w.Add(1)
	go func() {
		cb()
		w.Done()
	}()
}
