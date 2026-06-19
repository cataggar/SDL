const std = @import("std");
const skip_names = @import("skip.zig").names;

const c = @cImport({
    @cDefine("SDL_MAIN_HANDLED", "1");
    @cInclude("SDL3/SDL.h");
});

// translate-c text (names preserved) used to generate function wrappers.
const tc_text: []const u8 = @embedFile("cimport.zig");

fn isSkipped(comptime name: []const u8) bool {
    inline for (skip_names) |s| {
        if (comptime std.mem.eql(u8, s, name)) return true;
    }
    inline for ([_][]const u8{ "SDL_GLContextState", "SDL_iconv_data_t" }) |s| {
        if (comptime std.mem.eql(u8, s, name)) return true;
    }
    return false;
}

fn goName(n: []const u8) []const u8 {
    if (std.mem.startsWith(u8, n, "SDLK_")) return n[3..]; // SDLK_SPACE -> K_SPACE
    return n[4..]; // strip "SDL_"
}

// ---- comptime alias emission ----------------------------------------------

fn aliasable(comptime T: type) bool {
    return switch (@typeInfo(T)) {
        .@"struct", .@"union", .@"enum", .@"opaque", .int, .float => true,
        .pointer => |p| @typeInfo(p.child) != .@"fn",
        .optional => |o| switch (@typeInfo(o.child)) {
            .pointer => |p| @typeInfo(p.child) != .@"fn",
            else => false,
        },
        else => false,
    };
}

// ---- runtime type mapping for function wrappers ----------------------------

const Kind = enum { void, boolean, scalar, cstring, direct, bytecast, unsupported };
const Map = struct {
    kind: Kind = .unsupported,
    go: []const u8 = "", // Go type
    cc: []const u8 = "", // C cast type (scalar/bytecast)
};

fn trim(s: []const u8) []const u8 {
    return std.mem.trim(u8, s, " \t");
}

fn cPrim(s: []const u8) ?[]const u8 {
    const t = [_]struct { z: []const u8, cc: []const u8 }{
        .{ .z = "c_int", .cc = "C.int" },          .{ .z = "c_uint", .cc = "C.uint" },
        .{ .z = "c_long", .cc = "C.long" },        .{ .z = "c_ulong", .cc = "C.ulong" },
        .{ .z = "c_longlong", .cc = "C.longlong" }, .{ .z = "c_ulonglong", .cc = "C.ulonglong" },
        .{ .z = "c_short", .cc = "C.short" },      .{ .z = "c_ushort", .cc = "C.ushort" },
        .{ .z = "c_char", .cc = "C.char" },        .{ .z = "bool", .cc = "C.bool" },
        .{ .z = "u8", .cc = "C.uint8_t" },         .{ .z = "i8", .cc = "C.int8_t" },
        .{ .z = "u16", .cc = "C.uint16_t" },       .{ .z = "i16", .cc = "C.int16_t" },
        .{ .z = "u32", .cc = "C.uint32_t" },       .{ .z = "i32", .cc = "C.int32_t" },
        .{ .z = "u64", .cc = "C.uint64_t" },       .{ .z = "i64", .cc = "C.int64_t" },
        .{ .z = "usize", .cc = "C.size_t" },       .{ .z = "isize", .cc = "C.ptrdiff_t" },
        .{ .z = "f32", .cc = "C.float" },          .{ .z = "f64", .cc = "C.double" },
        .{ .z = "Uint8", .cc = "C.Uint8" },        .{ .z = "Sint8", .cc = "C.Sint8" },
        .{ .z = "Uint16", .cc = "C.Uint16" },      .{ .z = "Sint16", .cc = "C.Sint16" },
        .{ .z = "Uint32", .cc = "C.Uint32" },      .{ .z = "Sint32", .cc = "C.Sint32" },
        .{ .z = "Uint64", .cc = "C.Uint64" },      .{ .z = "Sint64", .cc = "C.Sint64" },
    };
    for (t) |e| if (std.mem.eql(u8, e.z, s)) return e.cc;
    return null;
}

fn gPrim(s: []const u8) ?[]const u8 {
    const t = [_]struct { z: []const u8, go: []const u8 }{
        .{ .z = "c_int", .go = "int32" },      .{ .z = "c_uint", .go = "uint32" },
        .{ .z = "c_long", .go = "int32" },     .{ .z = "c_ulong", .go = "uint32" },
        .{ .z = "c_longlong", .go = "int64" }, .{ .z = "c_ulonglong", .go = "uint64" },
        .{ .z = "c_short", .go = "int16" },    .{ .z = "c_ushort", .go = "uint16" },
        .{ .z = "c_char", .go = "int8" },
        .{ .z = "u8", .go = "uint8" },   .{ .z = "i8", .go = "int8" },
        .{ .z = "u16", .go = "uint16" }, .{ .z = "i16", .go = "int16" },
        .{ .z = "u32", .go = "uint32" }, .{ .z = "i32", .go = "int32" },
        .{ .z = "u64", .go = "uint64" }, .{ .z = "i64", .go = "int64" },
        .{ .z = "usize", .go = "uint" }, .{ .z = "isize", .go = "int" },
        .{ .z = "f32", .go = "float32" }, .{ .z = "f64", .go = "float64" },
        .{ .z = "Uint8", .go = "uint8" },   .{ .z = "Sint8", .go = "int8" },
        .{ .z = "Uint16", .go = "uint16" }, .{ .z = "Sint16", .go = "int16" },
        .{ .z = "Uint32", .go = "uint32" }, .{ .z = "Sint32", .go = "int32" },
        .{ .z = "Uint64", .go = "uint64" }, .{ .z = "Sint64", .go = "int64" },
    };
    for (t) |e| if (std.mem.eql(u8, e.z, s)) return e.go;
    return null;
}

const FuncPtrSet = struct {
    names: [512][]const u8 = undefined,
    n: usize = 0,
    fn has(self: *const FuncPtrSet, name: []const u8) bool {
        for (self.names[0..self.n]) |x| if (std.mem.eql(u8, x, name)) return true;
        return false;
    }
};

fn stripConst(s: []const u8) []const u8 {
    if (std.mem.startsWith(u8, s, "const ")) return s[6..];
    return s;
}

// C type for a pointer element. translate-c lowers `char` to u8, so a u8 element
// is a char buffer (cgo *C.char), distinct from `Uint8` buffers.
fn ptrElemC(s: []const u8) ?[]const u8 {
    if (std.mem.eql(u8, s, "u8")) return "C.char";
    return cPrim(s);
}

fn mapType(s_in: []const u8, fps: *const FuncPtrSet) Map {
    const s = trim(s_in);
    if (std.mem.eql(u8, s, "void")) return .{ .kind = .void };
    if (std.mem.eql(u8, s, "bool")) return .{ .kind = .boolean, .go = "bool" };
    if (std.mem.eql(u8, s, "[*c]const u8")) return .{ .kind = .cstring, .go = "string" };

    if (std.mem.startsWith(u8, s, "[*c]")) {
        const inner = stripConst(s[4..]);
        if (std.mem.eql(u8, inner, "anyopaque")) return .{ .kind = .direct, .go = "unsafe.Pointer" };
        if (std.mem.startsWith(u8, inner, "SDL_")) {
            if (fps.has(inner)) return .{};
            return .{ .kind = .direct, .go = std.fmt.allocPrint(g_alloc, "*{s}", .{goName(inner)}) catch "" };
        }
        if (ptrElemC(inner)) |cc| return .{ .kind = .bytecast, .go = "unsafe.Pointer", .cc = cc };
        return .{};
    }
    if (std.mem.startsWith(u8, s, "?*") or std.mem.startsWith(u8, s, "*")) {
        const after = if (std.mem.startsWith(u8, s, "?*")) s[2..] else s[1..];
        const inner = stripConst(after);
        if (std.mem.eql(u8, inner, "anyopaque")) return .{ .kind = .direct, .go = "unsafe.Pointer" };
        if (std.mem.startsWith(u8, inner, "SDL_")) {
            if (fps.has(inner)) return .{};
            return .{ .kind = .direct, .go = std.fmt.allocPrint(g_alloc, "*{s}", .{goName(inner)}) catch "" };
        }
        if (ptrElemC(inner)) |cc| return .{ .kind = .bytecast, .go = "unsafe.Pointer", .cc = cc };
        return .{};
    }
    if (std.mem.startsWith(u8, s, "SDL_")) {
        if (fps.has(s)) return .{};
        return .{ .kind = .direct, .go = goName(s) };
    }
    if (cPrim(s)) |cc| {
        if (gPrim(s)) |g| return .{ .kind = .scalar, .go = g, .cc = cc };
    }
    return .{};
}

var g_alloc: std.mem.Allocator = undefined;

pub fn main() !void {
    @setEvalBranchQuota(50_000_000);

    var arena_impl = std.heap.ArenaAllocator.init(std.heap.page_allocator);
    defer arena_impl.deinit();
    g_alloc = arena_impl.allocator();

    const io = std.Io.Threaded.global_single_threaded.io();
    var buf: [1 << 16]u8 = undefined;
    var fw = std.Io.File.stdout().writer(io, &buf);
    const w = &fw.interface;

    try w.writeAll(
        \\// Code generated by gosdl3/gen.zig from SDL3 headers. DO NOT EDIT.
        \\
        \\package sdl3
        \\
        \\/*
        \\#define SDL_MAIN_HANDLED
        \\#include <SDL3/SDL.h>
        \\#include <stdlib.h>
        \\*/
        \\import "C"
        \\
        \\import "unsafe"
        \\
        \\var _ = unsafe.Pointer(nil)
        \\
        \\// --- constants ---
        \\
    );

    inline for (@typeInfo(c).@"struct".decls) |d| {
        if (comptime (std.mem.startsWith(u8, d.name, "SDL_") or
            std.mem.startsWith(u8, d.name, "SDLK_")) and !isSkipped(d.name))
        {
            const V = @field(c, d.name);
            if (@TypeOf(V) != type) {
                switch (@typeInfo(@TypeOf(V))) {
                    .int, .comptime_int => {
                        const v: i128 = V;
                        try w.print("const {s} = {d}\n", .{ goName(d.name), v });
                    },
                    else => {},
                }
            }
        }
    }

    try w.writeAll("\n// --- type aliases ---\n\n");

    inline for (@typeInfo(c).@"struct".decls) |d| {
        if (comptime std.mem.startsWith(u8, d.name, "SDL_") and !isSkipped(d.name)) {
            const V = @field(c, d.name);
            if (@TypeOf(V) == type and comptime aliasable(V)) {
                try w.print("type {s} = C.{s}\n", .{ goName(d.name), d.name });
            }
        }
    }

    try w.writeAll("\n// --- functions (parsed from translate-c) ---\n\n");

    // Pass 1: collect function-pointer typedef names (callbacks) to skip.
    var fps = FuncPtrSet{};
    {
        var it = std.mem.splitScalar(u8, tc_text, '\n');
        while (it.next()) |line| {
            if (std.mem.startsWith(u8, line, "pub const SDL_") and
                std.mem.indexOf(u8, line, "fn (") != null)
            {
                const rest = line["pub const ".len..];
                if (std.mem.indexOf(u8, rest, " = ")) |eq| {
                    if (fps.n < fps.names.len) {
                        fps.names[fps.n] = rest[0..eq];
                        fps.n += 1;
                    }
                }
            }
        }
    }

    // Pass 2: emit a wrapper per extern fn.
    var it = std.mem.splitScalar(u8, tc_text, '\n');
    while (it.next()) |line| {
        if (!std.mem.startsWith(u8, line, "pub extern fn SDL_")) continue;
        try emitFn(w, line, &fps);
    }

    try w.flush();
}

fn emitFn(w: anytype, line: []const u8, fps: *const FuncPtrSet) !void {
    const after = line["pub extern fn ".len..];
    const lp = std.mem.indexOfScalar(u8, after, '(') orelse return;
    const cname = after[0..lp];

    // char* vs unsigned-char* is indistinguishable in translate-c text; these
    // few HID functions take uchar* buffers, so skip them.
    inline for ([_][]const u8{
        "SDL_hid_get_feature_report", "SDL_hid_get_input_report",
        "SDL_hid_get_report_descriptor", "SDL_hid_read", "SDL_hid_read_timeout",
        "SDL_hid_send_feature_report", "SDL_hid_write",
    }) |s| {
        if (std.mem.eql(u8, s, cname)) return;
    }

    // find matching ')'
    var depth: usize = 0;
    var end: usize = 0;
    var i: usize = lp;
    while (i < after.len) : (i += 1) {
        if (after[i] == '(') depth += 1;
        if (after[i] == ')') {
            depth -= 1;
            if (depth == 0) {
                end = i;
                break;
            }
        }
    }
    if (end == 0) return;
    const params_str = after[lp + 1 .. end];
    var ret_str = trim(after[end + 1 ..]);
    if (std.mem.endsWith(u8, ret_str, ";")) ret_str = trim(ret_str[0 .. ret_str.len - 1]);

    // function-pointer params -> skip
    if (std.mem.indexOf(u8, params_str, "fn (") != null) return;

    var ptypes: [24][]const u8 = undefined;
    var np: usize = 0;
    if (trim(params_str).len != 0) {
        var pit = std.mem.splitSequence(u8, params_str, ", ");
        while (pit.next()) |p| {
            const colon = std.mem.indexOf(u8, p, ": ") orelse return;
            if (np >= ptypes.len) return;
            ptypes[np] = trim(p[colon + 2 ..]);
            np += 1;
        }
    }

    // map all; bail if anything unsupported
    var maps: [24]Map = undefined;
    for (ptypes[0..np], 0..) |t, k| {
        maps[k] = mapType(t, fps);
        if (maps[k].kind == .unsupported or maps[k].kind == .void) return;
    }
    const ret = mapType(ret_str, fps);
    if (ret.kind == .unsupported) return;

    // signature
    try w.print("func {s}(", .{goName(cname)});
    for (maps[0..np], 0..) |m, k| {
        if (k != 0) try w.writeAll(", ");
        try w.print("a{d} {s}", .{ k, m.go });
    }
    if (ret.kind == .void) {
        try w.writeAll(") {\n");
    } else {
        try w.print(") {s} {{\n", .{ret.go});
    }

    // pre: C strings
    for (maps[0..np], 0..) |m, k| {
        if (m.kind == .cstring) {
            try w.print("\tc{d} := C.CString(a{d})\n\tdefer C.free(unsafe.Pointer(c{d}))\n", .{ k, k, k });
        }
    }

    // call
    try w.writeAll("\t");
    switch (ret.kind) {
        .void => {},
        .boolean => try w.writeAll("return bool("),
        .cstring => try w.writeAll("return C.GoString("),
        .scalar => try w.print("return {s}(", .{ret.go}),
        .bytecast => try w.writeAll("return unsafe.Pointer("),
        else => try w.writeAll("return "),
    }
    try w.print("C.{s}(", .{cname});
    for (maps[0..np], 0..) |m, k| {
        if (k != 0) try w.writeAll(", ");
        switch (m.kind) {
            .scalar => try w.print("{s}(a{d})", .{ m.cc, k }),
            .boolean => try w.print("C.bool(a{d})", .{k}),
            .cstring => try w.print("c{d}", .{k}),
            .bytecast => try w.print("(*{s})(a{d})", .{ m.cc, k }),
            else => try w.print("a{d}", .{k}),
        }
    }
    try w.writeAll(")");
    switch (ret.kind) {
        .boolean, .cstring, .scalar, .bytecast => try w.writeAll(")"),
        else => {},
    }
    try w.writeAll("\n}\n");
}
