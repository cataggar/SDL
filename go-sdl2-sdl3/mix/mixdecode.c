#include "mixaudio.h"

// Decode an audio buffer to PCM. Stage 1 handles WAV natively via SDL3; the
// compressed formats (MP3/OGG/FLAC) are added by the vendored decoders below.

bool mixdecode(const Uint8 *data, size_t len, SDL_AudioSpec *spec, Uint8 **buf, Uint32 *outlen) {
    if (!data || len == 0) return false;

    // WAV via SDL3 (SDL_LoadWAV_IO closes the IOStream itself).
    SDL_IOStream *io = SDL_IOFromConstMem(data, len);
    if (io && SDL_LoadWAV_IO(io, true, spec, buf, outlen)) {
        return true;
    }

    if (mixdecode_compressed(data, len, spec, buf, outlen)) {
        return true;
    }
    return false;
}
