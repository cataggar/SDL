//go:build !sdl3static

package sdl

// Default linkage: dynamic SDL3 (SDL3.dll must sit beside the executable).
// Built with `zig build install_sdl -Dpreferred_linkage=dynamic` from the SDL
// repo root (this package lives in that repo).

// #cgo CFLAGS: -I${SRCDIR}/../../zig-out/include
// #cgo LDFLAGS: -L${SRCDIR}/../../zig-out/lib -lSDL3
import "C"
