#ifndef GOSDL2_SDL3_CBITS_H
#define GOSDL2_SDL3_CBITS_H

#define SDL_MAIN_HANDLED
#include <SDL3/SDL.h>
#include <stdlib.h>

/*
 * Field accessors for the SDL_Event union. Hand-written so the Go side never
 * depends on cgo struct-field name mangling (e.g. `type` -> `_type`) or on the
 * exact union member layout.
 */
Uint32 shim_ev_type(SDL_Event *e);

Uint8  shim_btn_button(SDL_Event *e);
bool   shim_btn_down(SDL_Event *e);
float  shim_btn_x(SDL_Event *e);
float  shim_btn_y(SDL_Event *e);
Uint32 shim_btn_windowID(SDL_Event *e);

float  shim_mot_x(SDL_Event *e);
float  shim_mot_y(SDL_Event *e);
float  shim_mot_xrel(SDL_Event *e);
float  shim_mot_yrel(SDL_Event *e);
Uint32 shim_mot_state(SDL_Event *e);
Uint32 shim_mot_windowID(SDL_Event *e);

float  shim_wheel_x(SDL_Event *e);
float  shim_wheel_y(SDL_Event *e);
Sint32 shim_wheel_ix(SDL_Event *e);
Sint32 shim_wheel_iy(SDL_Event *e);
Uint32 shim_wheel_windowID(SDL_Event *e);

Uint32 shim_key_key(SDL_Event *e);
Uint16 shim_key_mod(SDL_Event *e);
bool   shim_key_down(SDL_Event *e);
bool   shim_key_repeat(SDL_Event *e);
Uint32 shim_key_windowID(SDL_Event *e);

const char *shim_text_text(SDL_Event *e);
Uint32      shim_text_windowID(SDL_Event *e);

const char *shim_edit_text(SDL_Event *e);
Sint32      shim_edit_start(SDL_Event *e);
Sint32      shim_edit_length(SDL_Event *e);
Uint32      shim_edit_windowID(SDL_Event *e);

Sint32 shim_win_data1(SDL_Event *e);
Sint32 shim_win_data2(SDL_Event *e);
Uint32 shim_win_windowID(SDL_Event *e);

float  shim_finger_x(SDL_Event *e);
float  shim_finger_y(SDL_Event *e);
Sint64 shim_finger_id(SDL_Event *e);
Uint32 shim_finger_windowID(SDL_Event *e);

const char *shim_drop_data(SDL_Event *e);
Uint32      shim_drop_windowID(SDL_Event *e);

/*
 * Read renderer pixels into dst (converting to `format`). SDL3's
 * SDL_RenderReadPixels returns a new surface rather than filling a buffer, so
 * this wraps the convert-and-copy.
 */
bool shim_render_read_pixels(SDL_Renderer *r, const SDL_Rect *rect, Uint32 format, void *dst, int pitch);

/* Mouse button masks (macros, wrapped so cgo never folds the expression). */
Uint32 shim_button_lmask(void);
Uint32 shim_button_rmask(void);
Uint32 shim_button_mmask(void);

/* GL context helpers (SDL_GLContext is an opaque pointer; pass it as void*). */
void *shim_gl_create_context(SDL_Window *w);
bool  shim_gl_make_current(SDL_Window *w, void *ctx);
bool  shim_gl_destroy_context(void *ctx);

/* Push a bare user event with the given type code (used to wake the loop). */
bool shim_push_user_event(Uint32 type);

/*
 * Event-watch trampoline. goEventWatch is exported from Go (watch.go); declared
 * here so cbits.c can hand its address to SDL_AddEventWatch.
 */
extern bool goEventWatch(void *userdata, SDL_Event *event);
void shim_add_event_watch(void *userdata);
void shim_remove_event_watch(void *userdata);

#endif /* GOSDL2_SDL3_CBITS_H */
