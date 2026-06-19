// Package sdl is a drop-in replacement for the subset of github.com/veandco/go-sdl2/sdl
// consumed by go-gui's SDL renderer backend (gui/backend/sdl2), its OpenGL
// backend (gui/backend/gl), the sdlkey helper, and the go-glyph SDL backend —
// implemented directly on SDL3 via cgo.
//
// Only the symbols those backends reference are provided. It deliberately
// reshapes SDL3's split window-event model and float coordinates back into the
// SDL2-style API that go-gui expects, so the backend code compiles and runs
// unchanged.
package sdl

/*
#include "cbits.h"
*/
import "C"

import (
	"errors"
	"unsafe"
)

// --- core types ---

// Keycode is an SDL virtual key code (Uint32).
type Keycode = C.SDL_Keycode

// Keymod is a bitmask of key modifiers (Uint16).
type Keymod = C.SDL_Keymod

// GLattr identifies an OpenGL context attribute.
type GLattr = C.SDL_GLAttr

// SystemCursor identifies a built-in system cursor shape.
type SystemCursor = C.SDL_SystemCursor

// GLContext is an opaque OpenGL context handle.
type GLContext = unsafe.Pointer

// Window wraps an SDL3 window.
type Window struct{ w *C.SDL_Window }

// Cursor wraps an SDL3 cursor.
type Cursor struct{ c *C.SDL_Cursor }

// Rect is an integer rectangle.
type Rect struct{ X, Y, W, H int32 }

// --- init / window / GL flags ---

const (
	INIT_VIDEO  = uint32(C.SDL_INIT_VIDEO)
	INIT_EVENTS = uint32(C.SDL_INIT_EVENTS)
)

const (
	// WINDOW_SHOWN has no SDL3 equivalent (windows are shown by default).
	WINDOW_SHOWN         = uint32(0)
	WINDOW_OPENGL        = uint32(C.SDL_WINDOW_OPENGL)
	WINDOW_RESIZABLE     = uint32(C.SDL_WINDOW_RESIZABLE)
	WINDOW_ALLOW_HIGHDPI = uint32(C.SDL_WINDOW_HIGH_PIXEL_DENSITY)
)

// WINDOWPOS_CENTERED requests centered window placement.
const WINDOWPOS_CENTERED = int32(C.SDL_WINDOWPOS_CENTERED)

const (
	GL_CONTEXT_MAJOR_VERSION = GLattr(C.SDL_GL_CONTEXT_MAJOR_VERSION)
	GL_CONTEXT_MINOR_VERSION = GLattr(C.SDL_GL_CONTEXT_MINOR_VERSION)
	GL_CONTEXT_PROFILE_MASK  = GLattr(C.SDL_GL_CONTEXT_PROFILE_MASK)
	GL_CONTEXT_FLAGS         = GLattr(C.SDL_GL_CONTEXT_FLAGS)
	GL_DOUBLEBUFFER          = GLattr(C.SDL_GL_DOUBLEBUFFER)
	GL_STENCIL_SIZE          = GLattr(C.SDL_GL_STENCIL_SIZE)
)

// Attribute values (passed as the value argument to GLSetAttribute).
const (
	GL_CONTEXT_PROFILE_CORE            = int(C.SDL_GL_CONTEXT_PROFILE_CORE)
	GL_CONTEXT_FORWARD_COMPATIBLE_FLAG = int(C.SDL_GL_CONTEXT_FORWARD_COMPATIBLE_FLAG)
)

// --- system cursors (mapped from SDL2 names to SDL3 shapes) ---

const (
	SYSTEM_CURSOR_ARROW     = SystemCursor(C.SDL_SYSTEM_CURSOR_DEFAULT)
	SYSTEM_CURSOR_IBEAM     = SystemCursor(C.SDL_SYSTEM_CURSOR_TEXT)
	SYSTEM_CURSOR_CROSSHAIR = SystemCursor(C.SDL_SYSTEM_CURSOR_CROSSHAIR)
	SYSTEM_CURSOR_HAND      = SystemCursor(C.SDL_SYSTEM_CURSOR_POINTER)
	SYSTEM_CURSOR_SIZEWE    = SystemCursor(C.SDL_SYSTEM_CURSOR_EW_RESIZE)
	SYSTEM_CURSOR_SIZENS    = SystemCursor(C.SDL_SYSTEM_CURSOR_NS_RESIZE)
	SYSTEM_CURSOR_SIZENWSE  = SystemCursor(C.SDL_SYSTEM_CURSOR_NWSE_RESIZE)
	SYSTEM_CURSOR_SIZENESW  = SystemCursor(C.SDL_SYSTEM_CURSOR_NESW_RESIZE)
	SYSTEM_CURSOR_SIZEALL   = SystemCursor(C.SDL_SYSTEM_CURSOR_MOVE)
	SYSTEM_CURSOR_NO        = SystemCursor(C.SDL_SYSTEM_CURSOR_NOT_ALLOWED)
)

// --- mouse buttons ---

const (
	BUTTON_LEFT   = 1
	BUTTON_MIDDLE = 2
	BUTTON_RIGHT  = 3
)

// ButtonLMask returns the left mouse button mask bit.
func ButtonLMask() uint32 { return uint32(C.shim_button_lmask()) }

// ButtonRMask returns the right mouse button mask bit.
func ButtonRMask() uint32 { return uint32(C.shim_button_rmask()) }

// ButtonMMask returns the middle mouse button mask bit.
func ButtonMMask() uint32 { return uint32(C.shim_button_mmask()) }

// --- key modifiers (combined left/right masks) ---

const (
	KMOD_SHIFT = Keymod(C.SDL_KMOD_SHIFT)
	KMOD_CTRL  = Keymod(C.SDL_KMOD_CTRL)
	KMOD_ALT   = Keymod(C.SDL_KMOD_ALT)
	KMOD_GUI   = Keymod(C.SDL_KMOD_GUI)
)

// --- keycodes ---

const (
	K_SPACE        = Keycode(C.SDLK_SPACE)
	K_RETURN       = Keycode(C.SDLK_RETURN)
	K_RETURN2      = Keycode(C.SDLK_RETURN2)
	K_KP_ENTER     = Keycode(C.SDLK_KP_ENTER)
	K_ESCAPE       = Keycode(C.SDLK_ESCAPE)
	K_TAB          = Keycode(C.SDLK_TAB)
	K_BACKSPACE    = Keycode(C.SDLK_BACKSPACE)
	K_DELETE       = Keycode(C.SDLK_DELETE)
	K_INSERT       = Keycode(C.SDLK_INSERT)
	K_RIGHT        = Keycode(C.SDLK_RIGHT)
	K_LEFT         = Keycode(C.SDLK_LEFT)
	K_DOWN         = Keycode(C.SDLK_DOWN)
	K_UP           = Keycode(C.SDLK_UP)
	K_PAGEUP       = Keycode(C.SDLK_PAGEUP)
	K_PAGEDOWN     = Keycode(C.SDLK_PAGEDOWN)
	K_HOME         = Keycode(C.SDLK_HOME)
	K_END          = Keycode(C.SDLK_END)
	K_LSHIFT       = Keycode(C.SDLK_LSHIFT)
	K_RSHIFT       = Keycode(C.SDLK_RSHIFT)
	K_LCTRL        = Keycode(C.SDLK_LCTRL)
	K_RCTRL        = Keycode(C.SDLK_RCTRL)
	K_LALT         = Keycode(C.SDLK_LALT)
	K_RALT         = Keycode(C.SDLK_RALT)
	K_LGUI         = Keycode(C.SDLK_LGUI)
	K_RGUI         = Keycode(C.SDLK_RGUI)
	K_COMMA        = Keycode(C.SDLK_COMMA)
	K_MINUS        = Keycode(C.SDLK_MINUS)
	K_PERIOD       = Keycode(C.SDLK_PERIOD)
	K_SLASH        = Keycode(C.SDLK_SLASH)
	K_SEMICOLON    = Keycode(C.SDLK_SEMICOLON)
	K_EQUALS       = Keycode(C.SDLK_EQUALS)
	K_LEFTBRACKET  = Keycode(C.SDLK_LEFTBRACKET)
	K_BACKSLASH    = Keycode(C.SDLK_BACKSLASH)
	K_RIGHTBRACKET = Keycode(C.SDLK_RIGHTBRACKET)
	K_BACKQUOTE    = Keycode(C.SDLK_GRAVE)
	K_CAPSLOCK     = Keycode(C.SDLK_CAPSLOCK)
	K_F1           = Keycode(C.SDLK_F1)
	K_F2           = Keycode(C.SDLK_F2)
	K_F3           = Keycode(C.SDLK_F3)
	K_F4           = Keycode(C.SDLK_F4)
	K_F5           = Keycode(C.SDLK_F5)
	K_F6           = Keycode(C.SDLK_F6)
	K_F7           = Keycode(C.SDLK_F7)
	K_F8           = Keycode(C.SDLK_F8)
	K_F9           = Keycode(C.SDLK_F9)
	K_F10          = Keycode(C.SDLK_F10)
	K_F11          = Keycode(C.SDLK_F11)
	K_F12          = Keycode(C.SDLK_F12)
)

// --- event type tags (exposed where go-gui compares Event.Type) ---

const (
	MOUSEBUTTONDOWN = uint32(C.SDL_EVENT_MOUSE_BUTTON_DOWN)
	MOUSEBUTTONUP   = uint32(C.SDL_EVENT_MOUSE_BUTTON_UP)
	KEYDOWN         = uint32(C.SDL_EVENT_KEY_DOWN)
	KEYUP           = uint32(C.SDL_EVENT_KEY_UP)

	FINGERDOWN   = uint32(C.SDL_EVENT_FINGER_DOWN)
	FINGERUP     = uint32(C.SDL_EVENT_FINGER_UP)
	FINGERMOTION = uint32(C.SDL_EVENT_FINGER_MOTION)
	DROPFILE     = uint32(C.SDL_EVENT_DROP_FILE)
)

// Synthetic SDL2-style window event subtypes. SDL3 splits these into distinct
// top-level event types; convertEvent folds them back into a single WindowEvent
// carrying one of these subtype tags.
const (
	WINDOWEVENT_NONE uint32 = iota
	WINDOWEVENT_RESIZED
	WINDOWEVENT_SIZE_CHANGED
	WINDOWEVENT_FOCUS_GAINED
	WINDOWEVENT_FOCUS_LOST
	WINDOWEVENT_CLOSE
)

// --- events ---

// Event is any SDL event. go-gui only type-switches over concrete pointers.
type Event interface{}

// Keysym holds the symbolic key and active modifiers for a key event.
type Keysym struct {
	Sym Keycode
	Mod uint16
}

type QuitEvent struct{}

type CommonEvent struct {
	Type     uint32
	WindowID uint32
}

type KeyboardEvent struct {
	Type     uint32
	WindowID uint32
	Keysym   Keysym
	Repeat   uint8
}

type MouseButtonEvent struct {
	Type     uint32
	WindowID uint32
	Button   uint8
	X, Y     int32
}

type MouseMotionEvent struct {
	WindowID         uint32
	X, Y, XRel, YRel int32
	State            uint32
}

type MouseWheelEvent struct {
	WindowID           uint32
	X, Y               int32
	PreciseX, PreciseY float32
}

type TextInputEvent struct {
	WindowID uint32
	text     string
}

// GetText returns the input text captured at poll time.
func (e *TextInputEvent) GetText() string { return e.text }

type TextEditingEvent struct {
	WindowID      uint32
	Start, Length int32
	text          string
}

// GetText returns the editing text captured at poll time.
func (e *TextEditingEvent) GetText() string { return e.text }

type WindowEvent struct {
	Event        uint32
	WindowID     uint32
	Data1, Data2 int32
}

type UserEvent struct {
	Type uint32
}

// TouchFingerEvent reports a touch-screen finger event. Coordinates are
// normalized 0..1 (as in SDL2).
type TouchFingerEvent struct {
	Type     uint32
	WindowID uint32
	FingerID int64
	X, Y     float32
}

// DropEvent reports a file drag-and-drop onto the window.
type DropEvent struct {
	Type     uint32
	WindowID uint32
	File     string
}

// --- error helper ---

// GetError returns the last SDL error string.
func GetError() string { return C.GoString(C.SDL_GetError()) }

func sdlError(context string) error {
	if msg := GetError(); msg != "" {
		return errors.New(context + ": " + msg)
	}
	return errors.New(context)
}

// --- lifecycle ---

// Init initializes the given SDL subsystems.
func Init(flags uint32) error {
	if !bool(C.SDL_Init(C.SDL_InitFlags(flags))) {
		return sdlError("SDL_Init")
	}
	return nil
}

// Quit shuts down all SDL subsystems.
func Quit() { C.SDL_Quit() }

// --- window ---

// CreateWindow creates an SDL3 window. The x/y arguments are honored via
// SDL_SetWindowPosition (SDL3's create call omits position).
func CreateWindow(title string, x, y, w, h int32, flags uint32) (*Window, error) {
	ctitle := C.CString(title)
	defer C.free(unsafe.Pointer(ctitle))
	win := C.SDL_CreateWindow(ctitle, C.int(w), C.int(h), C.SDL_WindowFlags(flags))
	if win == nil {
		return nil, sdlError("SDL_CreateWindow")
	}
	C.SDL_SetWindowPosition(win, C.int(x), C.int(y))
	return &Window{w: win}, nil
}

// GLCreateContext creates an OpenGL context for the window.
func (win *Window) GLCreateContext() (GLContext, error) {
	ctx := C.shim_gl_create_context(win.w)
	if ctx == nil {
		return nil, sdlError("SDL_GL_CreateContext")
	}
	return GLContext(ctx), nil
}

// GLMakeCurrent makes ctx current on this window.
func (win *Window) GLMakeCurrent(ctx GLContext) error {
	if !bool(C.shim_gl_make_current(win.w, unsafe.Pointer(ctx))) {
		return sdlError("SDL_GL_MakeCurrent")
	}
	return nil
}

// GLSwap swaps the window's OpenGL front and back buffers.
func (win *Window) GLSwap() { C.SDL_GL_SwapWindow(win.w) }

// GLGetDrawableSize returns the window's drawable size in pixels.
func (win *Window) GLGetDrawableSize() (int32, int32) {
	var pw, ph C.int
	C.SDL_GetWindowSizeInPixels(win.w, &pw, &ph)
	return int32(pw), int32(ph)
}

// GetSize returns the window's size in logical units.
func (win *Window) GetSize() (int32, int32) {
	var pw, ph C.int
	C.SDL_GetWindowSize(win.w, &pw, &ph)
	return int32(pw), int32(ph)
}

// GetID returns the window's numeric id.
func (win *Window) GetID() (uint32, error) {
	return uint32(C.SDL_GetWindowID(win.w)), nil
}

// SetTitle sets the window title.
func (win *Window) SetTitle(title string) {
	ctitle := C.CString(title)
	defer C.free(unsafe.Pointer(ctitle))
	C.SDL_SetWindowTitle(win.w, ctitle)
}

// Destroy destroys the window.
func (win *Window) Destroy() error {
	C.SDL_DestroyWindow(win.w)
	return nil
}

// --- GL attributes / context ---

// GLSetAttribute sets an OpenGL context creation attribute.
func GLSetAttribute(attr GLattr, value int) error {
	if !bool(C.SDL_GL_SetAttribute(attr, C.int(value))) {
		return sdlError("SDL_GL_SetAttribute")
	}
	return nil
}

// GLSetSwapInterval sets the OpenGL swap interval (vsync).
func GLSetSwapInterval(interval int) error {
	if !bool(C.SDL_GL_SetSwapInterval(C.int(interval))) {
		return sdlError("SDL_GL_SetSwapInterval")
	}
	return nil
}

// GLDeleteContext destroys an OpenGL context.
func GLDeleteContext(ctx GLContext) { C.shim_gl_destroy_context(unsafe.Pointer(ctx)) }

// --- cursors ---

// CreateSystemCursor creates a built-in system cursor.
func CreateSystemCursor(id SystemCursor) *Cursor {
	c := C.SDL_CreateSystemCursor(id)
	if c == nil {
		return nil
	}
	return &Cursor{c: c}
}

// SetCursor activates the given cursor.
func SetCursor(c *Cursor) {
	if c != nil {
		C.SDL_SetCursor(c.c)
	}
}

// FreeCursor frees a cursor.
func FreeCursor(c *Cursor) {
	if c != nil {
		C.SDL_DestroyCursor(c.c)
	}
}

// --- clipboard ---

// GetClipboardText returns the current clipboard text.
func GetClipboardText() (string, error) {
	cs := C.SDL_GetClipboardText()
	if cs == nil {
		return "", sdlError("SDL_GetClipboardText")
	}
	defer C.SDL_free(unsafe.Pointer(cs))
	return C.GoString(cs), nil
}

// SetClipboardText sets the clipboard text.
func SetClipboardText(text string) error {
	cs := C.CString(text)
	defer C.free(unsafe.Pointer(cs))
	if !bool(C.SDL_SetClipboardText(cs)) {
		return sdlError("SDL_SetClipboardText")
	}
	return nil
}

// --- keyboard / mouse state ---

// GetKeyboardFocus returns the window with keyboard focus, or nil.
func GetKeyboardFocus() *Window {
	w := C.SDL_GetKeyboardFocus()
	if w == nil {
		return nil
	}
	return &Window{w: w}
}

// GetModState returns the current key modifier state.
func GetModState() Keymod { return C.SDL_GetModState() }

// GetMouseState returns the cursor position and button state.
func GetMouseState() (int32, int32, uint32) {
	var x, y C.float
	st := C.SDL_GetMouseState(&x, &y)
	return int32(x), int32(y), uint32(st)
}

// --- text input (IME) ---

// StartTextInput enables text input for the focused window.
func StartTextInput() {
	if w := C.SDL_GetKeyboardFocus(); w != nil {
		C.SDL_StartTextInput(w)
	}
}

// StopTextInput disables text input for the focused window.
func StopTextInput() {
	if w := C.SDL_GetKeyboardFocus(); w != nil {
		C.SDL_StopTextInput(w)
	}
}

// SetTextInputRect sets the IME candidate area for the focused window.
func SetTextInputRect(r *Rect) {
	w := C.SDL_GetKeyboardFocus()
	if w == nil || r == nil {
		return
	}
	cr := C.SDL_Rect{x: C.int(r.X), y: C.int(r.Y), w: C.int(r.W), h: C.int(r.H)}
	C.SDL_SetTextInputArea(w, &cr, 0)
}

// --- event queue ---

// PollEvent returns the next pending event, or nil if the queue is empty.
func PollEvent() Event {
	var ev C.SDL_Event
	if !bool(C.SDL_PollEvent(&ev)) {
		return nil
	}
	return convertEvent(&ev)
}

// WaitEventTimeout waits up to timeout milliseconds for an event.
func WaitEventTimeout(timeout int) Event {
	var ev C.SDL_Event
	if !bool(C.SDL_WaitEventTimeout(&ev, C.Sint32(timeout))) {
		return nil
	}
	return convertEvent(&ev)
}

// PushEvent pushes a user event onto the queue.
func PushEvent(e *UserEvent) (bool, error) {
	if e == nil {
		return false, errors.New("nil event")
	}
	if !bool(C.shim_push_user_event(C.Uint32(e.Type))) {
		return false, sdlError("SDL_PushEvent")
	}
	return true, nil
}

// RegisterEvents allocates count user event type codes and returns the first.
func RegisterEvents(count int) uint32 { return uint32(C.SDL_RegisterEvents(C.int(count))) }

// convertEvent translates an SDL3 event union into the SDL2-style Go event
// types go-gui expects. Unmapped events become a *CommonEvent so the caller's
// drain loop keeps going (returning nil would prematurely end it).
func convertEvent(ce *C.SDL_Event) Event {
	t := uint32(C.shim_ev_type(ce))
	switch t {
	case uint32(C.SDL_EVENT_QUIT):
		return &QuitEvent{}

	case uint32(C.SDL_EVENT_MOUSE_BUTTON_DOWN), uint32(C.SDL_EVENT_MOUSE_BUTTON_UP):
		typ := MOUSEBUTTONUP
		if bool(C.shim_btn_down(ce)) {
			typ = MOUSEBUTTONDOWN
		}
		return &MouseButtonEvent{
			Type:     typ,
			WindowID: uint32(C.shim_btn_windowID(ce)),
			Button:   uint8(C.shim_btn_button(ce)),
			X:        int32(C.shim_btn_x(ce)),
			Y:        int32(C.shim_btn_y(ce)),
		}

	case uint32(C.SDL_EVENT_MOUSE_MOTION):
		return &MouseMotionEvent{
			WindowID: uint32(C.shim_mot_windowID(ce)),
			X:        int32(C.shim_mot_x(ce)),
			Y:        int32(C.shim_mot_y(ce)),
			XRel:     int32(C.shim_mot_xrel(ce)),
			YRel:     int32(C.shim_mot_yrel(ce)),
			State:    uint32(C.shim_mot_state(ce)),
		}

	case uint32(C.SDL_EVENT_MOUSE_WHEEL):
		return &MouseWheelEvent{
			WindowID: uint32(C.shim_wheel_windowID(ce)),
			X:        int32(C.shim_wheel_ix(ce)),
			Y:        int32(C.shim_wheel_iy(ce)),
			PreciseX: float32(C.shim_wheel_x(ce)),
			PreciseY: float32(C.shim_wheel_y(ce)),
		}

	case uint32(C.SDL_EVENT_KEY_DOWN), uint32(C.SDL_EVENT_KEY_UP):
		typ := KEYUP
		if bool(C.shim_key_down(ce)) {
			typ = KEYDOWN
		}
		var rep uint8
		if bool(C.shim_key_repeat(ce)) {
			rep = 1
		}
		return &KeyboardEvent{
			Type:     typ,
			WindowID: uint32(C.shim_key_windowID(ce)),
			Keysym: Keysym{
				Sym: Keycode(C.shim_key_key(ce)),
				Mod: uint16(C.shim_key_mod(ce)),
			},
			Repeat: rep,
		}

	case uint32(C.SDL_EVENT_TEXT_INPUT):
		return &TextInputEvent{
			WindowID: uint32(C.shim_text_windowID(ce)),
			text:     C.GoString(C.shim_text_text(ce)),
		}

	case uint32(C.SDL_EVENT_TEXT_EDITING):
		return &TextEditingEvent{
			WindowID: uint32(C.shim_edit_windowID(ce)),
			Start:    int32(C.shim_edit_start(ce)),
			Length:   int32(C.shim_edit_length(ce)),
			text:     C.GoString(C.shim_edit_text(ce)),
		}

	case uint32(C.SDL_EVENT_FINGER_DOWN), uint32(C.SDL_EVENT_FINGER_UP), uint32(C.SDL_EVENT_FINGER_MOTION):
		return &TouchFingerEvent{
			Type:     t,
			WindowID: uint32(C.shim_finger_windowID(ce)),
			FingerID: int64(C.shim_finger_id(ce)),
			X:        float32(C.shim_finger_x(ce)),
			Y:        float32(C.shim_finger_y(ce)),
		}

	case uint32(C.SDL_EVENT_DROP_FILE):
		return &DropEvent{
			Type:     DROPFILE,
			WindowID: uint32(C.shim_drop_windowID(ce)),
			File:     C.GoString(C.shim_drop_data(ce)),
		}

	case uint32(C.SDL_EVENT_WINDOW_RESIZED):
		return windowEvent(ce, WINDOWEVENT_RESIZED)
	case uint32(C.SDL_EVENT_WINDOW_PIXEL_SIZE_CHANGED):
		return windowEvent(ce, WINDOWEVENT_SIZE_CHANGED)
	case uint32(C.SDL_EVENT_WINDOW_FOCUS_GAINED):
		return windowEvent(ce, WINDOWEVENT_FOCUS_GAINED)
	case uint32(C.SDL_EVENT_WINDOW_FOCUS_LOST):
		return windowEvent(ce, WINDOWEVENT_FOCUS_LOST)
	case uint32(C.SDL_EVENT_WINDOW_CLOSE_REQUESTED):
		return windowEvent(ce, WINDOWEVENT_CLOSE)

	default:
		if t >= uint32(C.SDL_EVENT_USER) {
			return &UserEvent{Type: t}
		}
		return &CommonEvent{Type: t}
	}
}

func windowEvent(ce *C.SDL_Event, subtype uint32) *WindowEvent {
	return &WindowEvent{
		Event:    subtype,
		WindowID: uint32(C.shim_win_windowID(ce)),
		Data1:    int32(C.shim_win_data1(ce)),
		Data2:    int32(C.shim_win_data2(ce)),
	}
}
