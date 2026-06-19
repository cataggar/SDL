package sdl

/*
#include "cbits.h"
*/
import "C"

import "unsafe"

// --- renderer / texture / surface types ---

// Renderer wraps an SDL3 2D rendering context.
type Renderer struct{ r *C.SDL_Renderer }

// Texture wraps an SDL3 texture.
type Texture struct{ t *C.SDL_Texture }

// Surface wraps an SDL3 surface.
type Surface struct{ s *C.SDL_Surface }

// BlendMode selects how draws are blended.
type BlendMode = C.SDL_BlendMode

// Color is an 8-bit-per-channel RGBA color.
type Color struct{ R, G, B, A uint8 }

// FPoint is a float 2D point.
type FPoint struct{ X, Y float32 }

// FRect is a float rectangle.
type FRect struct{ X, Y, W, H float32 }

// Vertex is a geometry vertex (color is 8-bit RGBA as in go-sdl2; converted to
// SDL3's float SDL_FColor at draw time).
type Vertex struct {
	Position FPoint
	Color    Color
	TexCoord FPoint
}

// --- constants ---

const BLENDMODE_BLEND = BlendMode(C.SDL_BLENDMODE_BLEND)

// Renderer creation flags. SDL3 drops SDL2's flag set; only PRESENTVSYNC is
// acted upon (via SDL_SetRenderVSync). The values are arbitrary but distinct.
const (
	RENDERER_ACCELERATED  = uint32(0x00000002)
	RENDERER_PRESENTVSYNC = uint32(0x00000004)
)

const (
	PIXELFORMAT_RGBA8888 = uint32(C.SDL_PIXELFORMAT_RGBA8888)
	PIXELFORMAT_ABGR8888 = uint32(C.SDL_PIXELFORMAT_ABGR8888)
	PIXELFORMAT_RGBA32   = uint32(C.SDL_PIXELFORMAT_RGBA32)
)

const (
	TEXTUREACCESS_STATIC    = int(C.SDL_TEXTUREACCESS_STATIC)
	TEXTUREACCESS_STREAMING = int(C.SDL_TEXTUREACCESS_STREAMING)
	TEXTUREACCESS_TARGET    = int(C.SDL_TEXTUREACCESS_TARGET)
)

// HINT_RENDER_SCALE_QUALITY is accepted for API parity; SDL3 ignores it (its
// textures default to linear scaling).
const HINT_RENDER_SCALE_QUALITY = "SDL_RENDER_SCALE_QUALITY"

// --- helpers ---

func crect(r *Rect) *C.SDL_Rect {
	if r == nil {
		return nil
	}
	return &C.SDL_Rect{x: C.int(r.X), y: C.int(r.Y), w: C.int(r.W), h: C.int(r.H)}
}

func cfrect(r *FRect) *C.SDL_FRect {
	if r == nil {
		return nil
	}
	return &C.SDL_FRect{x: C.float(r.X), y: C.float(r.Y), w: C.float(r.W), h: C.float(r.H)}
}

func rectToFRect(r *Rect) *C.SDL_FRect {
	if r == nil {
		return nil
	}
	return &C.SDL_FRect{x: C.float(r.X), y: C.float(r.Y), w: C.float(r.W), h: C.float(r.H)}
}

// --- hints ---

// SetHint sets an SDL configuration hint.
func SetHint(name, value string) bool {
	cn := C.CString(name)
	defer C.free(unsafe.Pointer(cn))
	cv := C.CString(value)
	defer C.free(unsafe.Pointer(cv))
	return bool(C.SDL_SetHint(cn, cv))
}

// --- renderer ---

// CreateRenderer creates a 2D rendering context for the window. The index
// argument is ignored (SDL3 selects the driver, falling back to software on
// GPU-less hosts); RENDERER_PRESENTVSYNC enables vsync.
func CreateRenderer(win *Window, _ int, flags uint32) (*Renderer, error) {
	r := C.SDL_CreateRenderer(win.w, nil)
	if r == nil {
		return nil, sdlError("SDL_CreateRenderer")
	}
	if flags&RENDERER_PRESENTVSYNC != 0 {
		C.SDL_SetRenderVSync(r, 1)
	}
	return &Renderer{r: r}, nil
}

// SetDrawBlendMode sets the blend mode for primitive draws.
func (r *Renderer) SetDrawBlendMode(m BlendMode) error {
	if !bool(C.SDL_SetRenderDrawBlendMode(r.r, m)) {
		return sdlError("SDL_SetRenderDrawBlendMode")
	}
	return nil
}

// SetDrawColor sets the color for primitive draws.
func (r *Renderer) SetDrawColor(cr, cg, cb, ca uint8) error {
	if !bool(C.SDL_SetRenderDrawColor(r.r, C.Uint8(cr), C.Uint8(cg), C.Uint8(cb), C.Uint8(ca))) {
		return sdlError("SDL_SetRenderDrawColor")
	}
	return nil
}

// Clear clears the current render target with the draw color.
func (r *Renderer) Clear() error {
	if !bool(C.SDL_RenderClear(r.r)) {
		return sdlError("SDL_RenderClear")
	}
	return nil
}

// Present presents the rendered frame.
func (r *Renderer) Present() { C.SDL_RenderPresent(r.r) }

// Destroy destroys the renderer.
func (r *Renderer) Destroy() error {
	C.SDL_DestroyRenderer(r.r)
	return nil
}

// GetOutputSize returns the current render output size in pixels.
func (r *Renderer) GetOutputSize() (int32, int32, error) {
	var w, h C.int
	if !bool(C.SDL_GetCurrentRenderOutputSize(r.r, &w, &h)) {
		return 0, 0, sdlError("SDL_GetCurrentRenderOutputSize")
	}
	return int32(w), int32(h), nil
}

// SetClipRect sets the clip rectangle, or disables clipping when rect is nil.
func (r *Renderer) SetClipRect(rect *Rect) error {
	if !bool(C.SDL_SetRenderClipRect(r.r, crect(rect))) {
		return sdlError("SDL_SetRenderClipRect")
	}
	return nil
}

// GetClipRect returns the current clip rectangle.
func (r *Renderer) GetClipRect() Rect {
	var cr C.SDL_Rect
	C.SDL_GetRenderClipRect(r.r, &cr)
	return Rect{X: int32(cr.x), Y: int32(cr.y), W: int32(cr.w), H: int32(cr.h)}
}

// IsClipEnabled reports whether clipping is enabled.
func (r *Renderer) IsClipEnabled() bool { return bool(C.SDL_RenderClipEnabled(r.r)) }

// CreateTexture creates a texture with the given pixel format and access mode.
func (r *Renderer) CreateTexture(format uint32, access int, w, h int32) (*Texture, error) {
	t := C.SDL_CreateTexture(r.r, C.SDL_PixelFormat(format), C.SDL_TextureAccess(access), C.int(w), C.int(h))
	if t == nil {
		return nil, sdlError("SDL_CreateTexture")
	}
	return &Texture{t: t}, nil
}

// SetRenderTarget directs rendering to t, or to the window when t is nil.
func (r *Renderer) SetRenderTarget(t *Texture) error {
	var ct *C.SDL_Texture
	if t != nil {
		ct = t.t
	}
	if !bool(C.SDL_SetRenderTarget(r.r, ct)) {
		return sdlError("SDL_SetRenderTarget")
	}
	return nil
}

// GetRenderTarget returns the current render target texture, or nil for the
// window.
func (r *Renderer) GetRenderTarget() *Texture {
	t := C.SDL_GetRenderTarget(r.r)
	if t == nil {
		return nil
	}
	return &Texture{t: t}
}

// FillRect fills an integer rectangle.
func (r *Renderer) FillRect(rect *Rect) error {
	if !bool(C.SDL_RenderFillRect(r.r, rectToFRect(rect))) {
		return sdlError("SDL_RenderFillRect")
	}
	return nil
}

// FillRectF fills a float rectangle.
func (r *Renderer) FillRectF(rect *FRect) error {
	if !bool(C.SDL_RenderFillRect(r.r, cfrect(rect))) {
		return sdlError("SDL_RenderFillRect")
	}
	return nil
}

// DrawRectF strokes a float rectangle outline.
func (r *Renderer) DrawRectF(rect *FRect) error {
	if !bool(C.SDL_RenderRect(r.r, cfrect(rect))) {
		return sdlError("SDL_RenderRect")
	}
	return nil
}

// DrawLineF draws a line between two float points.
func (r *Renderer) DrawLineF(x1, y1, x2, y2 float32) error {
	if !bool(C.SDL_RenderLine(r.r, C.float(x1), C.float(y1), C.float(x2), C.float(y2))) {
		return sdlError("SDL_RenderLine")
	}
	return nil
}

// DrawPointF draws a single float point.
func (r *Renderer) DrawPointF(x, y float32) error {
	if !bool(C.SDL_RenderPoint(r.r, C.float(x), C.float(y))) {
		return sdlError("SDL_RenderPoint")
	}
	return nil
}

// Copy copies a texture region (integer src/dst rects).
func (r *Renderer) Copy(t *Texture, src, dst *Rect) error {
	return r.renderTexture(t, rectToFRect(src), rectToFRect(dst))
}

// CopyF copies a texture region (integer src, float dst).
func (r *Renderer) CopyF(t *Texture, src *Rect, dst *FRect) error {
	return r.renderTexture(t, rectToFRect(src), cfrect(dst))
}

func (r *Renderer) renderTexture(t *Texture, src, dst *C.SDL_FRect) error {
	var ct *C.SDL_Texture
	if t != nil {
		ct = t.t
	}
	if !bool(C.SDL_RenderTexture(r.r, ct, src, dst)) {
		return sdlError("SDL_RenderTexture")
	}
	return nil
}

// ReadPixels reads the current render target's pixels into the buffer at
// pixels, converting to the requested format.
func (r *Renderer) ReadPixels(rect *Rect, format uint32, pixels unsafe.Pointer, pitch int) error {
	if !bool(C.shim_render_read_pixels(r.r, crect(rect), C.Uint32(format), pixels, C.int(pitch))) {
		return sdlError("SDL_RenderReadPixels")
	}
	return nil
}

// RenderGeometry draws indexed (or sequential) triangles, optionally textured.
func (r *Renderer) RenderGeometry(t *Texture, vertices []Vertex, indices []int32) error {
	if len(vertices) == 0 {
		return nil
	}
	cverts := make([]C.SDL_Vertex, len(vertices))
	for i := range vertices {
		v := &vertices[i]
		cverts[i].position = C.SDL_FPoint{x: C.float(v.Position.X), y: C.float(v.Position.Y)}
		cverts[i].color = C.SDL_FColor{
			r: C.float(v.Color.R) / 255,
			g: C.float(v.Color.G) / 255,
			b: C.float(v.Color.B) / 255,
			a: C.float(v.Color.A) / 255,
		}
		cverts[i].tex_coord = C.SDL_FPoint{x: C.float(v.TexCoord.X), y: C.float(v.TexCoord.Y)}
	}
	var ct *C.SDL_Texture
	if t != nil {
		ct = t.t
	}
	var ip *C.int
	var ni C.int
	if len(indices) > 0 {
		ip = (*C.int)(unsafe.Pointer(&indices[0]))
		ni = C.int(len(indices))
	}
	if !bool(C.SDL_RenderGeometry(r.r, ct, &cverts[0], C.int(len(cverts)), ip, ni)) {
		return sdlError("SDL_RenderGeometry")
	}
	return nil
}

// --- texture ---

// SetBlendMode sets the texture's blend mode.
func (t *Texture) SetBlendMode(m BlendMode) error {
	if !bool(C.SDL_SetTextureBlendMode(t.t, m)) {
		return sdlError("SDL_SetTextureBlendMode")
	}
	return nil
}

// SetColorMod sets an additional color multiplier.
func (t *Texture) SetColorMod(r, g, b uint8) error {
	if !bool(C.SDL_SetTextureColorMod(t.t, C.Uint8(r), C.Uint8(g), C.Uint8(b))) {
		return sdlError("SDL_SetTextureColorMod")
	}
	return nil
}

// SetAlphaMod sets an additional alpha multiplier.
func (t *Texture) SetAlphaMod(a uint8) error {
	if !bool(C.SDL_SetTextureAlphaMod(t.t, C.Uint8(a))) {
		return sdlError("SDL_SetTextureAlphaMod")
	}
	return nil
}

// Update uploads pixel data to the texture region (whole texture when rect is
// nil).
func (t *Texture) Update(rect *Rect, pixels unsafe.Pointer, pitch int) error {
	if !bool(C.SDL_UpdateTexture(t.t, crect(rect), pixels, C.int(pitch))) {
		return sdlError("SDL_UpdateTexture")
	}
	return nil
}

// Query returns the texture's width and height. Format and access are not
// reported by SDL3's size API and are returned as zero.
func (t *Texture) Query() (format uint32, access int, w, h int32, err error) {
	var fw, fh C.float
	if !bool(C.SDL_GetTextureSize(t.t, &fw, &fh)) {
		return 0, 0, 0, 0, sdlError("SDL_GetTextureSize")
	}
	return 0, 0, int32(fw), int32(fh), nil
}

// Destroy destroys the texture.
func (t *Texture) Destroy() error {
	C.SDL_DestroyTexture(t.t)
	return nil
}

// --- surface ---

// CreateRGBSurfaceFrom wraps existing pixel data in a surface. The depth and
// channel masks are translated to an SDL3 pixel format. The pixel memory must
// outlive the surface.
func CreateRGBSurfaceFrom(pixels unsafe.Pointer, width, height int32, depth, pitch int, rmask, gmask, bmask, amask uint32) (*Surface, error) {
	format := C.SDL_GetPixelFormatForMasks(C.int(depth), C.Uint32(rmask), C.Uint32(gmask), C.Uint32(bmask), C.Uint32(amask))
	s := C.SDL_CreateSurfaceFrom(C.int(width), C.int(height), format, pixels, C.int(pitch))
	if s == nil {
		return nil, sdlError("SDL_CreateSurfaceFrom")
	}
	return &Surface{s: s}, nil
}

// Free frees the surface.
func (s *Surface) Free() {
	if s != nil {
		C.SDL_DestroySurface(s.s)
	}
}

// SetIcon sets the window icon from a surface.
func (win *Window) SetIcon(s *Surface) {
	if s != nil {
		C.SDL_SetWindowIcon(win.w, s.s)
	}
}
