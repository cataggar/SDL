// Package mix is a drop-in replacement for the subset of
// github.com/veandco/go-sdl2/mix used by go-gui's gui/audio package, backed by
// a small SDL3 audio mixing engine (see mixaudio.c) rather than SDL_mixer —
// which SDL3 does not bundle.
//
// Sound effects play on numbered channels (auto-mixed by SDL3); one music
// track plays at a time. WAV is decoded natively by SDL3; MP3/OGG/FLAC via
// vendored decoders.
package mix

/*
#include <stdlib.h>
#include "mixaudio.h"
*/
import "C"

import (
	"errors"
	"unsafe"

	"github.com/veandco/go-sdl2/sdl"
)

// Volume range and decoder-init flags, matching SDL_mixer's values.
const (
	MAX_VOLUME = 128

	// DEFAULT_FORMAT is SDL_AUDIO_S16 (used as uint16 by callers).
	DEFAULT_FORMAT = 0x8010

	INIT_FLAC = 0x01
	INIT_MOD  = 0x02
	INIT_MP3  = 0x08
	INIT_OGG  = 0x10
)

// Chunk is a loaded sound effect (decoded PCM).
type Chunk struct{ c *C.MixChunk }

// Music is a loaded music track (decoded PCM).
type Music struct{ m *C.MixMusic }

// Init is a no-op: decoders are compiled in, so there is nothing to load.
func Init(_ int) error { return nil }

// OpenAudio opens the output device. format and chunksize are accepted for API
// compatibility; SDL3 negotiates the device format and buffer size.
func OpenAudio(frequency int, _ uint16, channels, _ int) error {
	if C.mixeng_open(C.int(frequency), C.int(channels)) == 0 {
		return errors.New("mix: OpenAudio: " + sdl.GetError())
	}
	return nil
}

// AllocateChannels sets the number of mixing channels and returns it.
func AllocateChannels(numchans int) int {
	C.mixeng_allocate_channels(C.int(numchans))
	return numchans
}

// CloseAudio closes the output device and frees channels.
func CloseAudio() { C.mixeng_close() }

// Quit is a no-op (decoders are compiled in).
func Quit() {}

// LoadWAV loads a sound effect from a file (WAV/MP3/OGG/FLAC).
func LoadWAV(file string) (*Chunk, error) {
	cs := C.CString(file)
	defer C.free(unsafe.Pointer(cs))
	c := C.mixeng_load_file(cs)
	if c == nil {
		return nil, errors.New("mix: LoadWAV " + file + ": " + sdl.GetError())
	}
	return &Chunk{c: c}, nil
}

// LoadWAVRW loads a sound effect from an SDL_IOStream.
func LoadWAVRW(src *sdl.RWops, freesrc bool) (*Chunk, error) {
	if src == nil {
		return nil, errors.New("mix: LoadWAVRW: nil source")
	}
	var closeio C.int
	if freesrc {
		closeio = 1
	}
	c := C.mixeng_load_io((*C.SDL_IOStream)(src.Cptr()), closeio)
	if c == nil {
		return nil, errors.New("mix: LoadWAVRW: decode failed")
	}
	return &Chunk{c: c}, nil
}

// Play plays the chunk on channel (-1 = first free), with loops extra repeats
// (0 = once, -1 = forever). Returns the channel used.
func (c *Chunk) Play(channel, loops int) (int, error) {
	ch := int(C.mixeng_play(C.int(channel), c.c, C.int(loops), 0))
	if ch < 0 {
		return -1, errors.New("mix: Play: no free channel")
	}
	return ch, nil
}

// FadeIn is like Play but ramps volume up over ms milliseconds.
func (c *Chunk) FadeIn(channel, loops, ms int) (int, error) {
	ch := int(C.mixeng_play(C.int(channel), c.c, C.int(loops), C.int(ms)))
	if ch < 0 {
		return -1, errors.New("mix: FadeIn: no free channel")
	}
	return ch, nil
}

// Volume sets the chunk volume (0..128, -1 to query) and returns the previous.
func (c *Chunk) Volume(volume int) int {
	return int(C.mixeng_chunk_volume(c.c, C.int(volume)))
}

// Free releases the chunk.
func (c *Chunk) Free() {
	if c == nil || c.c == nil {
		return
	}
	C.mixeng_free_chunk(c.c)
	c.c = nil
}

// Volume sets a channel's volume (channel -1 = all), returning the previous.
func Volume(channel, volume int) int {
	return int(C.mixeng_channel_volume(C.int(channel), C.int(volume)))
}

// HaltChannel stops a channel (-1 = all).
func HaltChannel(channel int) int {
	C.mixeng_halt(C.int(channel))
	return 0
}

// FadeOutChannel fades out a channel over ms milliseconds (-1 = all).
func FadeOutChannel(channel, ms int) int {
	C.mixeng_fadeout(C.int(channel), C.int(ms))
	return 0
}

// Pause pauses a channel (-1 = all).
func Pause(channel int) { C.mixeng_pause(C.int(channel)) }

// Resume resumes a channel (-1 = all).
func Resume(channel int) { C.mixeng_resume(C.int(channel)) }

// Playing reports the number of playing channels (channel -1) or 1/0 for a
// specific channel.
func Playing(channel int) int { return int(C.mixeng_playing(C.int(channel))) }

// LoadMUS loads a music track from a file.
func LoadMUS(file string) (*Music, error) {
	cs := C.CString(file)
	defer C.free(unsafe.Pointer(cs))
	m := C.mixeng_load_music_file(cs)
	if m == nil {
		return nil, errors.New("mix: LoadMUS " + file + ": " + sdl.GetError())
	}
	return &Music{m: m}, nil
}

// Play starts the music with loops extra repeats (0 = once, -1 = forever),
// halting any current music first.
func (m *Music) Play(loops int) error {
	if C.mixeng_play_music(m.m, C.int(loops), 0) == 0 {
		return errors.New("mix: play music: " + sdl.GetError())
	}
	return nil
}

// FadeIn is like Play but ramps up over ms milliseconds.
func (m *Music) FadeIn(loops, ms int) error {
	if C.mixeng_play_music(m.m, C.int(loops), C.int(ms)) == 0 {
		return errors.New("mix: fade-in music: " + sdl.GetError())
	}
	return nil
}

// Free releases the music track.
func (m *Music) Free() {
	if m == nil || m.m == nil {
		return
	}
	C.mixeng_free_music(m.m)
	m.m = nil
}

// HaltMusic stops the current music.
func HaltMusic() error { C.mixeng_halt_music(); return nil }

// FadeOutMusic fades the music out over ms milliseconds.
func FadeOutMusic(ms int) bool { C.mixeng_fadeout_music(C.int(ms)); return true }

// PauseMusic pauses the music.
func PauseMusic() { C.mixeng_pause_music() }

// ResumeMusic resumes paused music.
func ResumeMusic() { C.mixeng_resume_music() }

// RewindMusic restarts the current music from the beginning.
func RewindMusic() { C.mixeng_rewind_music() }

// PlayingMusic reports whether music is playing (including while paused).
func PlayingMusic() bool { return C.mixeng_playing_music() != 0 }

// PausedMusic reports whether the music is paused.
func PausedMusic() bool { return C.mixeng_paused_music() != 0 }

// VolumeMusic sets the music volume (0..128, -1 to query), returning previous.
func VolumeMusic(volume int) int { return int(C.mixeng_music_volume(C.int(volume))) }
