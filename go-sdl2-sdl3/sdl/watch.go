package sdl

/*
#include "cbits.h"
*/
import "C"

import (
	"sync"
	"unsafe"
)

// EventWatchHandle identifies a registered event watch.
type EventWatchHandle int

// EventWatchCallback is invoked for every event while a watch is active.
// The bool result is accepted for API parity with go-sdl2 but ignored, as
// SDL event watches cannot veto events.
type EventWatchCallback func(Event, any) bool

var (
	watchMu         sync.Mutex
	watchSeq        int
	watches         = map[EventWatchHandle]EventWatchCallback{}
	watchRegistered bool
)

// AddEventWatchFunc registers fn to be called for every event, including those
// delivered during modal OS loops (e.g. live window resize). The data argument
// is accepted for API compatibility and passed through as nil to fn.
//
// The single C trampoline is registered with SDL only once; all Go callbacks
// are dispatched from the registry, so handles never need to be smuggled
// through SDL's void* userdata.
func AddEventWatchFunc(fn EventWatchCallback, _ any) EventWatchHandle {
	watchMu.Lock()
	watchSeq++
	h := EventWatchHandle(watchSeq)
	watches[h] = fn
	register := !watchRegistered
	watchRegistered = true
	watchMu.Unlock()
	if register {
		C.shim_add_event_watch(nil)
	}
	return h
}

// DelEventWatch removes a previously registered event watch. When the last
// watch is removed, the C trampoline is detached from SDL.
func DelEventWatch(h EventWatchHandle) {
	watchMu.Lock()
	delete(watches, h)
	empty := len(watches) == 0
	if empty {
		watchRegistered = false
	}
	watchMu.Unlock()
	if empty {
		C.shim_remove_event_watch(nil)
	}
}

//export goEventWatch
func goEventWatch(_ unsafe.Pointer, cev *C.SDL_Event) C.bool {
	watchMu.Lock()
	if len(watches) == 0 {
		watchMu.Unlock()
		return C.bool(true)
	}
	fns := make([]EventWatchCallback, 0, len(watches))
	for _, fn := range watches {
		fns = append(fns, fn)
	}
	watchMu.Unlock()

	ev := convertEvent(cev)
	for _, fn := range fns {
		fn(ev, nil)
	}
	return C.bool(true)
}
