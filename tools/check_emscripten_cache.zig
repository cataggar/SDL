const std = @import("std");

pub fn main(init: std.process.Init) !void {
    const args = try init.minimal.args.toSlice(init.arena.allocator());
    if (args.len != 2) {
        std.log.err("expected the Emscripten cache include directory", .{});
        return error.InvalidArguments;
    }
    var dir = std.Io.Dir.cwd().openDir(init.io, args[1], .{
        .access_sub_paths = true,
        .follow_symlinks = false,
    }) catch |err| {
        std.log.err("cannot open Emscripten cache include directory: {s}", .{@errorName(err)});
        return err;
    };
    dir.close(init.io);
}
