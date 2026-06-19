//go:build !sdl3static

package mix

// Default linkage: dynamic SDL3 (SDL3.dll must sit beside the executable).

// #cgo CFLAGS: -I${SRCDIR}/../../zig-out/include
// #cgo LDFLAGS: -L${SRCDIR}/../../zig-out/lib -lSDL3
import "C"
