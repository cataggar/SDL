#include "cbits.h"

Uint32 shim_ev_type(SDL_Event *e) { return e->type; }

Uint8  shim_btn_button(SDL_Event *e) { return e->button.button; }
bool   shim_btn_down(SDL_Event *e) { return e->button.down; }
float  shim_btn_x(SDL_Event *e) { return e->button.x; }
float  shim_btn_y(SDL_Event *e) { return e->button.y; }
Uint32 shim_btn_windowID(SDL_Event *e) { return (Uint32)e->button.windowID; }

float  shim_mot_x(SDL_Event *e) { return e->motion.x; }
float  shim_mot_y(SDL_Event *e) { return e->motion.y; }
float  shim_mot_xrel(SDL_Event *e) { return e->motion.xrel; }
float  shim_mot_yrel(SDL_Event *e) { return e->motion.yrel; }
Uint32 shim_mot_state(SDL_Event *e) { return (Uint32)e->motion.state; }
Uint32 shim_mot_windowID(SDL_Event *e) { return (Uint32)e->motion.windowID; }

float  shim_wheel_x(SDL_Event *e) { return e->wheel.x; }
float  shim_wheel_y(SDL_Event *e) { return e->wheel.y; }
Sint32 shim_wheel_ix(SDL_Event *e) { return e->wheel.integer_x; }
Sint32 shim_wheel_iy(SDL_Event *e) { return e->wheel.integer_y; }
Uint32 shim_wheel_windowID(SDL_Event *e) { return (Uint32)e->wheel.windowID; }

Uint32 shim_key_key(SDL_Event *e) { return (Uint32)e->key.key; }
Uint16 shim_key_mod(SDL_Event *e) { return (Uint16)e->key.mod; }
bool   shim_key_down(SDL_Event *e) { return e->key.down; }
bool   shim_key_repeat(SDL_Event *e) { return e->key.repeat; }
Uint32 shim_key_windowID(SDL_Event *e) { return (Uint32)e->key.windowID; }

const char *shim_text_text(SDL_Event *e) { return e->text.text; }
Uint32      shim_text_windowID(SDL_Event *e) { return (Uint32)e->text.windowID; }

const char *shim_edit_text(SDL_Event *e) { return e->edit.text; }
Sint32      shim_edit_start(SDL_Event *e) { return e->edit.start; }
Sint32      shim_edit_length(SDL_Event *e) { return e->edit.length; }
Uint32      shim_edit_windowID(SDL_Event *e) { return (Uint32)e->edit.windowID; }

Sint32 shim_win_data1(SDL_Event *e) { return e->window.data1; }
Sint32 shim_win_data2(SDL_Event *e) { return e->window.data2; }
Uint32 shim_win_windowID(SDL_Event *e) { return (Uint32)e->window.windowID; }

float  shim_finger_x(SDL_Event *e) { return e->tfinger.x; }
float  shim_finger_y(SDL_Event *e) { return e->tfinger.y; }
Sint64 shim_finger_id(SDL_Event *e) { return (Sint64)e->tfinger.fingerID; }
Uint32 shim_finger_windowID(SDL_Event *e) { return (Uint32)e->tfinger.windowID; }

const char *shim_drop_data(SDL_Event *e) { return e->drop.data; }
Uint32      shim_drop_windowID(SDL_Event *e) { return (Uint32)e->drop.windowID; }

bool shim_render_read_pixels(SDL_Renderer *r, const SDL_Rect *rect, Uint32 format, void *dst, int pitch) {
    SDL_Surface *surf = SDL_RenderReadPixels(r, rect);
    if (!surf) {
        return false;
    }
    SDL_Surface *conv = surf;
    if ((Uint32)surf->format != format) {
        conv = SDL_ConvertSurface(surf, (SDL_PixelFormat)format);
        SDL_DestroySurface(surf);
        if (!conv) {
            return false;
        }
    }
    int rowbytes = pitch < conv->pitch ? pitch : conv->pitch;
    for (int y = 0; y < conv->h; y++) {
        SDL_memcpy((Uint8 *)dst + (size_t)y * pitch,
                   (Uint8 *)conv->pixels + (size_t)y * conv->pitch,
                   (size_t)rowbytes);
    }
    SDL_DestroySurface(conv);
    return true;
}

Uint32 shim_button_lmask(void) { return SDL_BUTTON_LMASK; }
Uint32 shim_button_rmask(void) { return SDL_BUTTON_RMASK; }
Uint32 shim_button_mmask(void) { return SDL_BUTTON_MMASK; }

void *shim_gl_create_context(SDL_Window *w) { return (void *)SDL_GL_CreateContext(w); }
bool  shim_gl_make_current(SDL_Window *w, void *ctx) { return SDL_GL_MakeCurrent(w, (SDL_GLContext)ctx); }
bool  shim_gl_destroy_context(void *ctx) { return SDL_GL_DestroyContext((SDL_GLContext)ctx); }

bool shim_push_user_event(Uint32 type) {
    SDL_Event e;
    SDL_memset(&e, 0, sizeof(e));
    e.type = type;
    return SDL_PushEvent(&e);
}

void shim_add_event_watch(void *userdata) { SDL_AddEventWatch(goEventWatch, userdata); }
void shim_remove_event_watch(void *userdata) { SDL_RemoveEventWatch(goEventWatch, userdata); }
