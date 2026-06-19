package sdl

/*
#include "cbits.h"
*/
import "C"

import (
	"errors"
	"unsafe"
)

// INIT_AUDIO is the SDL audio subsystem flag.
const INIT_AUDIO = uint32(C.SDL_INIT_AUDIO)

// InitSubSystem initializes an additional SDL subsystem after SDL_Init.
func InitSubSystem(flags uint32) error {
	if !bool(C.SDL_InitSubSystem(C.SDL_InitFlags(flags))) {
		return sdlError("SDL_InitSubSystem")
	}
	return nil
}

// QuitSubSystem shuts down an SDL subsystem.
func QuitSubSystem(flags uint32) { C.SDL_QuitSubSystem(C.SDL_InitFlags(flags)) }

// RWops wraps an SDL3 SDL_IOStream. It exists so the go-sdl2 audio API
// (sdl.RWFromMem + mix.LoadWAVRW) keeps working on SDL3.
type RWops struct{ io unsafe.Pointer }

// RWFromMem creates a read-only SDL_IOStream over an in-memory buffer. The
// buffer must remain alive (and unmodified) while the RWops is in use.
func RWFromMem(mem []byte) (*RWops, error) {
	if len(mem) == 0 {
		return nil, errors.New("sdl: RWFromMem: empty buffer")
	}
	io := C.SDL_IOFromConstMem(unsafe.Pointer(&mem[0]), C.size_t(len(mem)))
	if io == nil {
		return nil, sdlError("SDL_IOFromConstMem")
	}
	return &RWops{io: unsafe.Pointer(io)}, nil
}

// Cptr returns the underlying SDL_IOStream pointer. Intended for the companion
// mix package, which casts it back to its own cgo SDL_IOStream type.
func (rw *RWops) Cptr() unsafe.Pointer {
	if rw == nil {
		return nil
	}
	return rw.io
}
