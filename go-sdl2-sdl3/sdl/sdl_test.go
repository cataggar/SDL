package sdl

import "testing"

// These tests exercise the shim's trickiest cgo paths (event push/poll/convert,
// the event-watch trampoline, clipboard round-trip) without needing OpenGL, so
// they pass on GPU-less hosts. They require a usable SDL video driver; if none
// is available the suite is skipped.

func mustInit(t *testing.T) {
	t.Helper()
	if err := Init(INIT_VIDEO | INIT_EVENTS); err != nil {
		t.Skipf("SDL video unavailable: %v", err)
	}
}

func TestUserEventRoundTrip(t *testing.T) {
	mustInit(t)
	defer Quit()

	base := RegisterEvents(1)
	if base == 0 || base == 0xFFFFFFFF {
		t.Fatalf("RegisterEvents returned invalid base %d", base)
	}

	var watched uint32
	h := AddEventWatchFunc(func(ev Event, _ any) bool {
		if ue, ok := ev.(*UserEvent); ok {
			watched = ue.Type
		}
		return true
	}, nil)
	defer DelEventWatch(h)

	ok, err := PushEvent(&UserEvent{Type: base})
	if err != nil || !ok {
		t.Fatalf("PushEvent: ok=%v err=%v", ok, err)
	}

	if watched != base {
		t.Errorf("event watch trampoline: got type %d, want %d", watched, base)
	}

	var found bool
	for i := 0; i < 256; i++ {
		ev := PollEvent()
		if ev == nil {
			break
		}
		if ue, ok := ev.(*UserEvent); ok && ue.Type == base {
			found = true
		}
	}
	if !found {
		t.Errorf("pushed user event not returned by PollEvent")
	}
}

func TestClipboardRoundTrip(t *testing.T) {
	mustInit(t)
	defer Quit()

	const want = "go-sdl2-sdl3-shim ☑"
	if err := SetClipboardText(want); err != nil {
		t.Fatalf("SetClipboardText: %v", err)
	}
	got, err := GetClipboardText()
	if err != nil {
		t.Fatalf("GetClipboardText: %v", err)
	}
	if got != want {
		t.Errorf("clipboard round-trip: got %q, want %q", got, want)
	}
}

func TestCursorsAndState(t *testing.T) {
	mustInit(t)
	defer Quit()

	for _, id := range []SystemCursor{
		SYSTEM_CURSOR_ARROW, SYSTEM_CURSOR_IBEAM, SYSTEM_CURSOR_HAND,
		SYSTEM_CURSOR_SIZEWE, SYSTEM_CURSOR_SIZENWSE, SYSTEM_CURSOR_NO,
	} {
		c := CreateSystemCursor(id)
		if c == nil {
			t.Errorf("CreateSystemCursor(%d) returned nil", id)
			continue
		}
		SetCursor(c)
		FreeCursor(c)
	}

	// Button masks must be the distinct SDL bits.
	if ButtonLMask() == 0 || ButtonRMask() == 0 || ButtonMMask() == 0 {
		t.Errorf("button masks zero: L=%d R=%d M=%d", ButtonLMask(), ButtonRMask(), ButtonMMask())
	}
	if ButtonLMask() == ButtonRMask() || ButtonLMask() == ButtonMMask() {
		t.Errorf("button masks not distinct: L=%d R=%d M=%d", ButtonLMask(), ButtonRMask(), ButtonMMask())
	}

	_, _, _ = GetMouseState()
	_ = GetModState()
}

func TestKeycodeConstantsDistinct(t *testing.T) {
	// K_RETURN2 must differ from K_RETURN so the sdlkey switch cases compile
	// and behave (a duplicate would silently fold).
	if K_RETURN == K_RETURN2 {
		t.Errorf("K_RETURN and K_RETURN2 collide (%d)", K_RETURN)
	}
	if K_RETURN == K_KP_ENTER {
		t.Errorf("K_RETURN and K_KP_ENTER collide (%d)", K_RETURN)
	}
}
