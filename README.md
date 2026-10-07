# SDL for Zig

This is a fork of [SDL](https://www.libsdl.org/), packaged for [Zig](https://ziglang.org).
Unnecessary files have been deleted, and the build system has been replaced with `build.zig`.
The package provides version 2 of SDL. For version 3, consider https://github.com/castholm/SDL.

This wrapper requires Zig 0.17.0 and retains SDL 2.32.10.

## Getting started

### Linking SDL2 to your project

Fetch SDL and add to your `build.zig.zon` :
```bash
zig fetch --save=SDL git+https://github.com/cataggar/SDL#compat/sdl2-2.32-zig017
```

Add this to your `build.zig` :
```zig
const sdl_dep = b.dependency("SDL", .{
    .optimize = .fast,
    .target = target,
});
exe.root_module.linkLibrary(sdl_dep.artifact("SDL2"));
```

### Emscripten

Pass `-Demscripten-root="$EMSDK/upstream/emscripten"` to select an initialized
Emscripten installation. A dependency supplies the same `emscripten-root`
option as a `std.Build.LazyPath`. The wrapper uses its cached sysroot headers
through a tracked lazy include path. A make-time check retains the directory
and no-follow validation; missing caches and symlinked include directories fail.

The compiler's global `--sysroot` flag is still supported by Zig, but is now a
make-time input and cannot be read during package configuration. Supply the
explicit package option in addition to any existing compiler `--sysroot` flag.
