#include "mixaudio.h"

// ---------------------------------------------------------------------------
// SDL3 audio mixing engine. One logical playback device; each sound-effect
// "channel" and the single music slot is an SDL_AudioStream bound to it (SDL
// mixes all bound streams). Looping/feeding happens in a pure-C get-callback.
// ---------------------------------------------------------------------------

#define MIX_MAX_VOL 128

struct MixChunk {
    Uint8        *buf;
    Uint32        len;
    SDL_AudioSpec spec;
    int           volume; // 0..128
};

struct MixMusic {
    Uint8        *buf;
    Uint32        len;
    SDL_AudioSpec spec;
};

typedef struct {
    SDL_AudioStream *stream;
    MixChunk        *chunk;  // current chunk (not owned)
    int              loops;  // remaining extra repeats after the first; -1 = forever
    int              volume; // 0..128 channel volume
    bool             active;
    bool             paused;
    int              fade_dir;   // +1 in, -1 out, 0 none
    Uint64           fade_start; // ms
    int              fade_ms;
} MixChannel;

static SDL_AudioDeviceID g_dev = 0;
static SDL_AudioSpec     g_devspec;
static MixChannel       *g_ch = NULL;
static int               g_nch = 0;

static SDL_AudioStream *g_mstream = NULL;
static MixMusic        *g_mcur = NULL; // not owned
static int    g_mloops = 0;
static int    g_mvol = MIX_MAX_VOL;
static bool   g_mactive = false;
static bool   g_mpaused = false;
static int    g_mfade_dir = 0;
static Uint64 g_mfade_start = 0;
static int    g_mfade_ms = 0;

static float fade_factor(int dir, Uint64 start, int ms) {
    if (dir == 0 || ms <= 0) return 1.0f;
    float t = (float)(SDL_GetTicks() - start) / (float)ms;
    if (t < 0) t = 0;
    if (t > 1) t = 1;
    return dir > 0 ? t : (1.0f - t);
}

static void apply_channel_gain(MixChannel *c) {
    float g = (c->volume / (float)MIX_MAX_VOL);
    if (c->chunk) g *= (c->chunk->volume / (float)MIX_MAX_VOL);
    g *= fade_factor(c->fade_dir, c->fade_start, c->fade_ms);
    SDL_SetAudioStreamGain(c->stream, g);
}

static void SDLCALL channel_cb(void *userdata, SDL_AudioStream *stream,
                               int additional, int total) {
    (void)total;
    MixChannel *c = (MixChannel *)userdata;
    if (additional <= 0 || !c->active || !c->chunk) return;

    if (c->fade_dir < 0 && c->fade_ms > 0 &&
        SDL_GetTicks() - c->fade_start >= (Uint64)c->fade_ms) {
        c->active = false;
        c->loops = 0;
        return; // fade-out complete; let the queue drain to silence
    }
    apply_channel_gain(c);

    if (c->loops == 0) return; // last iteration is draining
    SDL_PutAudioStreamData(stream, c->chunk->buf, (int)c->chunk->len);
    if (c->loops > 0) c->loops--;
}

static bool ch_is_playing(MixChannel *c) {
    if (!c->active) return false;
    if (c->loops != 0) return true;
    return SDL_GetAudioStreamQueued(c->stream) > 0;
}

int mixeng_open(int freq, int out_channels) {
    if (g_dev) return 1;
    SDL_AudioSpec spec;
    SDL_zero(spec);
    spec.freq = freq > 0 ? freq : 44100;
    spec.format = SDL_AUDIO_S16;
    spec.channels = out_channels > 0 ? out_channels : 2;
    g_devspec = spec;

    g_dev = SDL_OpenAudioDevice(SDL_AUDIO_DEVICE_DEFAULT_PLAYBACK, &spec);
    if (!g_dev) return 0;

    g_mstream = SDL_CreateAudioStream(&spec, &spec);
    if (g_mstream) {
        SDL_BindAudioStream(g_dev, g_mstream);
    }
    SDL_ResumeAudioDevice(g_dev);
    return 1;
}

void mixeng_allocate_channels(int n) {
    if (n < 0) n = 0;
    if (!g_dev) { g_nch = n; return; }

    for (int i = n; i < g_nch; i++) { // shrinking: destroy extras
        if (g_ch[i].stream) {
            SDL_UnbindAudioStream(g_ch[i].stream);
            SDL_DestroyAudioStream(g_ch[i].stream);
        }
    }
    MixChannel *nc = (MixChannel *)SDL_realloc(g_ch, sizeof(MixChannel) * (n > 0 ? n : 1));
    if (!nc && n > 0) return;
    g_ch = nc;
    for (int i = g_nch; i < n; i++) { // growing: create streams
        SDL_zero(g_ch[i]);
        g_ch[i].volume = MIX_MAX_VOL;
        g_ch[i].stream = SDL_CreateAudioStream(&g_devspec, &g_devspec);
        if (g_ch[i].stream) {
            SDL_SetAudioStreamGetCallback(g_ch[i].stream, channel_cb, &g_ch[i]);
            SDL_BindAudioStream(g_dev, g_ch[i].stream);
        }
    }
    g_nch = n;
}

void mixeng_close(void) {
    if (g_ch) {
        for (int i = 0; i < g_nch; i++) {
            if (g_ch[i].stream) {
                SDL_UnbindAudioStream(g_ch[i].stream);
                SDL_DestroyAudioStream(g_ch[i].stream);
            }
        }
        SDL_free(g_ch);
        g_ch = NULL;
    }
    g_nch = 0;
    if (g_mstream) {
        SDL_UnbindAudioStream(g_mstream);
        SDL_DestroyAudioStream(g_mstream);
        g_mstream = NULL;
    }
    g_mcur = NULL;
    g_mactive = false;
    g_mpaused = false;
    if (g_dev) {
        SDL_CloseAudioDevice(g_dev);
        g_dev = 0;
    }
}

static MixChunk *chunk_from_pcm(SDL_AudioSpec *spec, Uint8 *buf, Uint32 len) {
    MixChunk *c = (MixChunk *)SDL_calloc(1, sizeof(MixChunk));
    if (!c) { SDL_free(buf); return NULL; }
    c->buf = buf;
    c->len = len;
    c->spec = *spec;
    c->volume = MIX_MAX_VOL;
    return c;
}

MixChunk *mixeng_load_io(SDL_IOStream *io, int closeio) {
    if (!io) return NULL;
    size_t size = 0;
    void *data = SDL_LoadFile_IO(io, &size, closeio != 0);
    if (!data) return NULL;
    SDL_AudioSpec spec;
    Uint8 *buf = NULL;
    Uint32 len = 0;
    bool ok = mixdecode((const Uint8 *)data, size, &spec, &buf, &len);
    SDL_free(data);
    if (!ok) return NULL;
    return chunk_from_pcm(&spec, buf, len);
}

MixChunk *mixeng_load_file(const char *path) {
    if (!path || !path[0]) return NULL;
    size_t size = 0;
    void *data = SDL_LoadFile(path, &size);
    if (!data) return NULL;
    SDL_AudioSpec spec;
    Uint8 *buf = NULL;
    Uint32 len = 0;
    bool ok = mixdecode((const Uint8 *)data, size, &spec, &buf, &len);
    SDL_free(data);
    if (!ok) return NULL;
    return chunk_from_pcm(&spec, buf, len);
}

void mixeng_free_chunk(MixChunk *c) {
    if (!c) return;
    for (int i = 0; i < g_nch; i++) { // stop any channel using it
        if (g_ch[i].chunk == c) mixeng_halt(i);
    }
    SDL_free(c->buf);
    SDL_free(c);
}

int mixeng_chunk_volume(MixChunk *c, int volume) {
    if (!c) return 0;
    int prev = c->volume;
    if (volume >= 0) c->volume = volume > MIX_MAX_VOL ? MIX_MAX_VOL : volume;
    return prev;
}

int mixeng_play(int channel, MixChunk *c, int loops, int fade_ms) {
    if (!g_dev || !c || g_nch == 0) return -1;
    if (channel < 0) {
        for (int i = 0; i < g_nch; i++) {
            if (!ch_is_playing(&g_ch[i])) { channel = i; break; }
        }
        if (channel < 0) return -1;
    }
    if (channel >= g_nch || !g_ch[channel].stream) return -1;

    MixChannel *ch = &g_ch[channel];
    SDL_LockAudioStream(ch->stream);
    SDL_ClearAudioStream(ch->stream);
    ch->chunk = c;
    ch->loops = loops;
    ch->active = true;
    ch->paused = false;
    ch->fade_dir = fade_ms > 0 ? 1 : 0;
    ch->fade_start = SDL_GetTicks();
    ch->fade_ms = fade_ms;
    SDL_SetAudioStreamFormat(ch->stream, &c->spec, &g_devspec);
    apply_channel_gain(ch);
    SDL_PutAudioStreamData(ch->stream, c->buf, (int)c->len);
    SDL_UnlockAudioStream(ch->stream);
    SDL_BindAudioStream(g_dev, ch->stream); // no-op if already bound
    return channel;
}

void mixeng_halt(int channel) {
    if (!g_ch) return;
    if (channel < 0) {
        for (int i = 0; i < g_nch; i++) mixeng_halt(i);
        return;
    }
    if (channel >= g_nch || !g_ch[channel].stream) return;
    MixChannel *ch = &g_ch[channel];
    SDL_LockAudioStream(ch->stream);
    SDL_ClearAudioStream(ch->stream);
    ch->active = false;
    ch->loops = 0;
    ch->chunk = NULL;
    ch->paused = false;
    SDL_UnlockAudioStream(ch->stream);
    SDL_BindAudioStream(g_dev, ch->stream);
}

void mixeng_fadeout(int channel, int ms) {
    if (!g_ch) return;
    if (ms <= 0) { mixeng_halt(channel); return; }
    if (channel < 0) {
        for (int i = 0; i < g_nch; i++) mixeng_fadeout(i, ms);
        return;
    }
    if (channel >= g_nch) return;
    MixChannel *ch = &g_ch[channel];
    if (!ch->active) return;
    ch->fade_dir = -1;
    ch->fade_start = SDL_GetTicks();
    ch->fade_ms = ms;
}

void mixeng_pause(int channel) {
    if (!g_ch) return;
    if (channel < 0) {
        for (int i = 0; i < g_nch; i++) mixeng_pause(i);
        return;
    }
    if (channel >= g_nch || !g_ch[channel].stream) return;
    if (!g_ch[channel].paused) {
        g_ch[channel].paused = true;
        SDL_UnbindAudioStream(g_ch[channel].stream);
    }
}

void mixeng_resume(int channel) {
    if (!g_ch) return;
    if (channel < 0) {
        for (int i = 0; i < g_nch; i++) mixeng_resume(i);
        return;
    }
    if (channel >= g_nch || !g_ch[channel].stream) return;
    if (g_ch[channel].paused) {
        g_ch[channel].paused = false;
        SDL_BindAudioStream(g_dev, g_ch[channel].stream);
    }
}

int mixeng_playing(int channel) {
    if (!g_ch) return 0;
    if (channel < 0) {
        int n = 0;
        for (int i = 0; i < g_nch; i++) if (ch_is_playing(&g_ch[i])) n++;
        return n;
    }
    if (channel >= g_nch) return 0;
    return ch_is_playing(&g_ch[channel]) ? 1 : 0;
}

int mixeng_channel_volume(int channel, int volume) {
    if (!g_ch || g_nch == 0) return 0;
    if (channel < 0) {
        int prev = g_ch[0].volume;
        for (int i = 0; i < g_nch; i++) {
            if (volume >= 0) g_ch[i].volume = volume > MIX_MAX_VOL ? MIX_MAX_VOL : volume;
            apply_channel_gain(&g_ch[i]);
        }
        return prev;
    }
    if (channel >= g_nch) return 0;
    int prev = g_ch[channel].volume;
    if (volume >= 0) {
        g_ch[channel].volume = volume > MIX_MAX_VOL ? MIX_MAX_VOL : volume;
        apply_channel_gain(&g_ch[channel]);
    }
    return prev;
}

// --- music ---------------------------------------------------------------

static float music_gain(void) {
    float g = g_mvol / (float)MIX_MAX_VOL;
    g *= fade_factor(g_mfade_dir, g_mfade_start, g_mfade_ms);
    return g;
}

static void music_feed(void) {
    if (!g_mstream || !g_mcur) return;
    SDL_SetAudioStreamFormat(g_mstream, &g_mcur->spec, &g_devspec);
    SDL_SetAudioStreamGain(g_mstream, music_gain());
    SDL_PutAudioStreamData(g_mstream, g_mcur->buf, (int)g_mcur->len);
}

static void SDLCALL music_cb(void *userdata, SDL_AudioStream *stream,
                             int additional, int total) {
    (void)userdata; (void)stream; (void)total;
    if (additional <= 0 || !g_mactive || !g_mcur) return;
    if (g_mfade_dir < 0 && g_mfade_ms > 0 &&
        SDL_GetTicks() - g_mfade_start >= (Uint64)g_mfade_ms) {
        g_mactive = false;
        g_mloops = 0;
        return;
    }
    SDL_SetAudioStreamGain(g_mstream, music_gain());
    if (g_mloops == 0) return;
    SDL_PutAudioStreamData(g_mstream, g_mcur->buf, (int)g_mcur->len);
    if (g_mloops > 0) g_mloops--;
}

MixMusic *mixeng_load_music_file(const char *path) {
    if (!path || !path[0]) return NULL;
    size_t size = 0;
    void *data = SDL_LoadFile(path, &size);
    if (!data) return NULL;
    SDL_AudioSpec spec;
    Uint8 *buf = NULL;
    Uint32 len = 0;
    bool ok = mixdecode((const Uint8 *)data, size, &spec, &buf, &len);
    SDL_free(data);
    if (!ok) return NULL;
    MixMusic *m = (MixMusic *)SDL_calloc(1, sizeof(MixMusic));
    if (!m) { SDL_free(buf); return NULL; }
    m->buf = buf;
    m->len = len;
    m->spec = spec;
    return m;
}

void mixeng_free_music(MixMusic *m) {
    if (!m) return;
    if (g_mcur == m) mixeng_halt_music();
    SDL_free(m->buf);
    SDL_free(m);
}

int mixeng_play_music(MixMusic *m, int loops, int fade_ms) {
    if (!g_dev || !g_mstream || !m) return 0;
    SDL_LockAudioStream(g_mstream);
    SDL_ClearAudioStream(g_mstream);
    g_mcur = m;
    g_mloops = loops;
    g_mactive = true;
    g_mpaused = false;
    g_mfade_dir = fade_ms > 0 ? 1 : 0;
    g_mfade_start = SDL_GetTicks();
    g_mfade_ms = fade_ms;
    SDL_SetAudioStreamGetCallback(g_mstream, music_cb, NULL);
    music_feed();
    SDL_UnlockAudioStream(g_mstream);
    SDL_BindAudioStream(g_dev, g_mstream);
    return 1;
}

void mixeng_halt_music(void) {
    if (!g_mstream) return;
    SDL_LockAudioStream(g_mstream);
    SDL_ClearAudioStream(g_mstream);
    g_mactive = false;
    g_mloops = 0;
    g_mcur = NULL;
    g_mpaused = false;
    SDL_UnlockAudioStream(g_mstream);
    SDL_BindAudioStream(g_dev, g_mstream);
}

void mixeng_fadeout_music(int ms) {
    if (ms <= 0) { mixeng_halt_music(); return; }
    if (!g_mactive) return;
    g_mfade_dir = -1;
    g_mfade_start = SDL_GetTicks();
    g_mfade_ms = ms;
}

void mixeng_pause_music(void) {
    if (g_mstream && !g_mpaused && g_mactive) {
        g_mpaused = true;
        SDL_UnbindAudioStream(g_mstream);
    }
}

void mixeng_resume_music(void) {
    if (g_mstream && g_mpaused) {
        g_mpaused = false;
        SDL_BindAudioStream(g_dev, g_mstream);
    }
}

void mixeng_rewind_music(void) {
    if (!g_mstream || !g_mcur || !g_mactive) return;
    SDL_LockAudioStream(g_mstream);
    SDL_ClearAudioStream(g_mstream);
    music_feed();
    SDL_UnlockAudioStream(g_mstream);
}

int mixeng_playing_music(void) {
    if (!g_mactive) return 0;
    if (g_mpaused) return 1;
    if (g_mloops != 0) return 1;
    return (g_mstream && SDL_GetAudioStreamQueued(g_mstream) > 0) ? 1 : 0;
}

int mixeng_paused_music(void) { return g_mpaused ? 1 : 0; }

int mixeng_music_volume(int volume) {
    int prev = g_mvol;
    if (volume >= 0) {
        g_mvol = volume > MIX_MAX_VOL ? MIX_MAX_VOL : volume;
        if (g_mstream) SDL_SetAudioStreamGain(g_mstream, music_gain());
    }
    return prev;
}
