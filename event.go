package xterm

// Ported from xterm.js src/common/Event.ts.
// Synchronous, callback-based event system. No channels.

import "slices"

// EventEmitter is a synchronous event emitter.
// Listeners are called in registration order when Fire is called.
type EventEmitter[T any] struct {
	listeners []*eventEntry[T]
	disposed  bool
}

type eventEntry[T any] struct {
	fn      func(T)
	removed bool
}

// Fire invokes all registered listeners synchronously with the given value.
// Listeners that unsubscribe during iteration are handled safely via snapshot.
func (e *EventEmitter[T]) Fire(value T) {
	if e.disposed {
		return
	}
	// Unsubscribe never mutates the slice in place, so ranging over it is a
	// stable snapshot even if listeners change during iteration.
	ls := e.listeners
	if len(ls) == 1 {
		if !ls[0].removed {
			ls[0].fn(value)
		}
		return
	}
	for _, entry := range ls {
		if !entry.removed {
			entry.fn(value)
		}
	}
}

// Event registers a listener and returns a Disposable that removes it.
func (e *EventEmitter[T]) Event(listener func(T)) Disposable {
	if e.disposed {
		return nopDisposable
	}
	entry := &eventEntry[T]{fn: listener}
	e.listeners = append(e.listeners, entry)
	return toDisposable(func() {
		entry.removed = true
		for i, l := range e.listeners {
			if l == entry {
				e.listeners = slices.Delete(slices.Clone(e.listeners), i, i+1)
				break
			}
		}
	})
}

// Dispose removes all listeners and prevents future subscriptions.
func (e *EventEmitter[T]) Dispose() {
	if e.disposed {
		return
	}
	e.disposed = true
	e.listeners = nil
}

// HasListeners returns true if there are any active listeners.
func (e *EventEmitter[T]) HasListeners() bool {
	return len(e.listeners) > 0
}
