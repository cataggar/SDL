//go:build sdl3static

package sdl

// Static linkage (build with -tags sdl3static): SDL3 is linked into the binary,
// so no SDL3.dll is needed at runtime. Requires a static SDL3 built into a
// separate prefix, from the SDL repo root (this package lives in that repo):
//
//   zig build install_sdl -Dpreferred_linkage=static -Doptimize=ReleaseFast \
//     -Dstrip=true --prefix zig-out-static
//
// The Win32 system libraries below mirror SDL's own build.zig dependency list;
// a static archive cannot carry them, so consumers must re-link them.

// #cgo CFLAGS: -I${SRCDIR}/../../zig-out-static/include
// #cgo LDFLAGS: -L${SRCDIR}/../../zig-out-static/lib -lSDL3 -lkernel32 -luser32 -lgdi32 -lwinmm -limm32 -lole32 -loleaut32 -lversion -luuid -ladvapi32 -lsetupapi -lshell32 -ldinput8
import "C"
