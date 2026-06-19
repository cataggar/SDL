#ifndef GOSDL2_SDL3_MIXAUDIO_H
#define GOSDL2_SDL3_MIXAUDIO_H

#define SDL_MAIN_HANDLED
#include <SDL3/SDL.h>

// A minimal SDL_mixer-compatible audio engine built on SDL3's audio API:
// one logical playback device with N per-channel SDL_AudioStreams (SDL mixes
// all bound streams) plus a dedicated music stream. Looping is driven by a
// pure-C get-callback so no Go callback is involved.

typedef struct MixChunk MixChunk;
typedef struct MixMusic MixMusic;

int  mixeng_open(int freq, int out_channels);
void mixeng_close(void);
void mixeng_allocate_channels(int n);

// Chunk = a fully decoded sample buffer (WAV via SDL; compressed formats via
// the decoders in mixdecode.c).
MixChunk *mixeng_load_io(SDL_IOStream *io, int closeio);
MixChunk *mixeng_load_file(const char *path);
void      mixeng_free_chunk(MixChunk *c);
int       mixeng_chunk_volume(MixChunk *c, int volume);

int  mixeng_play(int channel, MixChunk *c, int loops, int fade_ms);
void mixeng_halt(int channel);
void mixeng_fadeout(int channel, int ms);
void mixeng_pause(int channel);
void mixeng_resume(int channel);
int  mixeng_playing(int channel);
int  mixeng_channel_volume(int channel, int volume);

MixMusic *mixeng_load_music_file(const char *path);
void      mixeng_free_music(MixMusic *m);
int       mixeng_play_music(MixMusic *m, int loops, int fade_ms);
void      mixeng_halt_music(void);
void      mixeng_fadeout_music(int ms);
void      mixeng_pause_music(void);
void      mixeng_resume_music(void);
void      mixeng_rewind_music(void);
int       mixeng_playing_music(void);
int       mixeng_paused_music(void);
int       mixeng_music_volume(int volume);

// Decode a compressed (or WAV) audio buffer to PCM. Implemented in
// mixdecode.c. Returns true and fills *spec/*buf/*outlen (buffer owned by
// caller, free with SDL_free) on success. Tries WAV, then MP3, OGG, FLAC.
bool mixdecode(const Uint8 *data, size_t len, SDL_AudioSpec *spec, Uint8 **buf, Uint32 *outlen);

// Decode MP3/OGG/FLAC via the vendored decoders (mixdecode_compressed.c).
bool mixdecode_compressed(const Uint8 *data, size_t len, SDL_AudioSpec *spec, Uint8 **buf, Uint32 *outlen);

#endif // GOSDL2_SDL3_MIXAUDIO_H
