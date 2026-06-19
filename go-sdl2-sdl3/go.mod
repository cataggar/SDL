// Module github.com/veandco/go-sdl2 is a drop-in replacement for the subset of
// the go-sdl2 `sdl` package API used by go-gui's SDL renderer and OpenGL
// backends (and the go-glyph SDL backend), implemented directly on top of SDL3
// (cgo). It exists so go-gui can target SDL3 without the sdl2-compat shim,
// enabling a single-DLL (or fully static) build.
module github.com/veandco/go-sdl2

go 1.26
