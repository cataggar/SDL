//go:build sdl3static

package mix

// Static linkage (build with -tags sdl3static): SDL3 is linked into the binary.
// See ../sdl/link_static.go for the matching SDL3 static build instructions.

// #cgo CFLAGS: -I${SRCDIR}/../../zig-out-static/include
// #cgo LDFLAGS: -L${SRCDIR}/../../zig-out-static/lib -lSDL3 -lkernel32 -luser32 -lgdi32 -lwinmm -limm32 -lole32 -loleaut32 -lversion -luuid -ladvapi32 -lsetupapi -lshell32 -ldinput8
import "C"
