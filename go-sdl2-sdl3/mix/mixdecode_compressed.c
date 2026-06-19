#include "mixaudio.h"
#include <stdlib.h>

// Declarations only (implementations live in dr_mp3_impl.c / stb_vorbis_impl.c).
#define DR_MP3_NO_STDIO
#include "dr_mp3.h"

#define STB_VORBIS_HEADER_ONLY
#define STB_VORBIS_NO_STDIO
#define STB_VORBIS_NO_PUSHDATA_API
#include "stb_vorbis.h"

static bool try_mp3(const Uint8 *data, size_t len, SDL_AudioSpec *spec,
                    Uint8 **buf, Uint32 *outlen) {
    drmp3_config cfg;
    drmp3_uint64 frames = 0;
    drmp3_int16 *pcm = drmp3_open_memory_and_read_pcm_frames_s16(
        data, len, &cfg, &frames, NULL);
    if (!pcm) return false;
    Uint32 bytes = (Uint32)(frames * cfg.channels * sizeof(drmp3_int16));
    Uint8 *out = (Uint8 *)SDL_malloc(bytes ? bytes : 1);
    if (!out) { drmp3_free(pcm, NULL); return false; }
    SDL_memcpy(out, pcm, bytes);
    drmp3_free(pcm, NULL);
    SDL_zero(*spec);
    spec->format = SDL_AUDIO_S16;
    spec->channels = (int)cfg.channels;
    spec->freq = (int)cfg.sampleRate;
    *buf = out;
    *outlen = bytes;
    return true;
}

static bool try_ogg(const Uint8 *data, size_t len, SDL_AudioSpec *spec,
                    Uint8 **buf, Uint32 *outlen) {
    int channels = 0, rate = 0;
    short *output = NULL;
    int samples = stb_vorbis_decode_memory(data, (int)len, &channels, &rate, &output);
    if (samples < 0 || !output) return false;
    Uint32 bytes = (Uint32)samples * (Uint32)channels * (Uint32)sizeof(short);
    Uint8 *out = (Uint8 *)SDL_malloc(bytes ? bytes : 1);
    if (!out) { free(output); return false; }
    SDL_memcpy(out, output, bytes);
    free(output); // stb_vorbis allocates the output with the C allocator
    SDL_zero(*spec);
    spec->format = SDL_AUDIO_S16;
    spec->channels = channels;
    spec->freq = rate;
    *buf = out;
    *outlen = bytes;
    return true;
}

bool mixdecode_compressed(const Uint8 *data, size_t len, SDL_AudioSpec *spec,
                          Uint8 **buf, Uint32 *outlen) {
    if (!data || len < 4) return false;

    // Route by magic, with a best-effort fallback.
    if (data[0] == 'O' && data[1] == 'g' && data[2] == 'g' && data[3] == 'S') {
        return try_ogg(data, len, spec, buf, outlen);
    }
    if ((data[0] == 'I' && data[1] == 'D' && data[2] == '3') ||
        (data[0] == 0xFF && (data[1] & 0xE0) == 0xE0)) {
        return try_mp3(data, len, spec, buf, outlen);
    }
    if (try_mp3(data, len, spec, buf, outlen)) return true;
    if (try_ogg(data, len, spec, buf, outlen)) return true;
    return false;
}
