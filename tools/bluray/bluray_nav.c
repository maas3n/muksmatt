/*
 * muKsMaTT Blu-ray navigation helper.
 * Compile against system libbluray: cc -O2 -Wall -Wextra -Werror -o muksmatt-bluray-nav
 *   tools/bluray/bluray_nav.c $(pkg-config --cflags --libs libbluray)
 *
 * stdout is a deliberately narrow versioned, tab-separated protocol:
 *   MUKSMATT_BD_NAV_1
 *   P<TAB>playlist-number<TAB>duration-90kHz<TAB>clip-count
 *   C<TAB>playlist-number<TAB>chapter-start-90kHz<TAB>chapter-end-90kHz
 *
 * The reader NEVER requests decryption keys or downloads AACS/BD+ data.
 * libbluray handles availability and protection checks according to its build
 * and user's existing lawful configuration. No M2TS contents are exported here.
 */
#include <inttypes.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <libbluray/bluray.h>

static int usage(void) {
    fputs("Usage: muksmatt-bluray-nav SOURCE\n       muksmatt-bluray-nav --stream-languages PLAYLIST SOURCE\n", stderr);
    return 2;
}

static int valid_lang(const uint8_t lang[4]) {
    if (lang[3] != 0) return 0;
    for (unsigned i = 0; i < 3; ++i) {
        if (!((lang[i] >= 'a' && lang[i] <= 'z') ||
              (lang[i] >= 'A' && lang[i] <= 'Z'))) return 0;
    }
    return 1;
}

static int print_stream_languages(const BLURAY_TITLE_INFO *title) {
    unsigned records = 0;
    puts("MUKSMATT_BD_LANG_1");
    for (uint32_t c = 0; c < title->clip_count; ++c) {
        const BLURAY_CLIP_INFO *clip = &title->clips[c];
        for (int category = 0; category < 2; ++category) {
            const BLURAY_STREAM_INFO *streams = category == 0 ? clip->audio_streams : clip->pg_streams;
            unsigned count = category == 0 ? clip->audio_stream_count : clip->pg_stream_count;
            if (count && !streams) return 1;
            for (unsigned i = 0; i < count; ++i) {
                if (++records > 16384) return 1;
                if (streams[i].pid == 0 || streams[i].pid >= 8192 || !valid_lang(streams[i].lang)) continue;
                printf("S\t%c\t%u\t%c%c%c\n", category == 0 ? 'a' : 's',
                       streams[i].pid, streams[i].lang[0], streams[i].lang[1], streams[i].lang[2]);
            }
        }
    }
    return 0;
}

int main(int argc, char **argv) {
    const int language_mode = argc == 4 && !strcmp(argv[1], "--stream-languages");
    if ((!language_mode && argc != 2) || !argv[1] || !argv[1][0]) return usage();
    if (!language_mode && !strcmp(argv[1], "--version")) {
        puts("MUKSMATT_BD_NAV_1");
        return 0;
    }
    unsigned requested_playlist = 0;
    if (language_mode) {
        if (!argv[2] || !*argv[2] || strlen(argv[2]) > 5) return usage();
        for (const char *p = argv[2]; *p; ++p) if (*p < '0' || *p > '9') return usage();
        requested_playlist = (unsigned)strtoul(argv[2], NULL, 10);
        if (requested_playlist > 99999 || !argv[3] || !*argv[3]) return usage();
    }
    BLURAY *disc = bd_open(language_mode ? argv[3] : argv[1], NULL);
    if (!disc) {
        fputs("libbluray could not open the Blu-ray source. Check image/device access and libudfread support.\n", stderr);
        return 1;
    }
    const BLURAY_DISC_INFO *info = bd_get_disc_info(disc);
    if (!info || (info->aacs_detected && !info->aacs_handled) ||
                 (info->bdplus_detected && !info->bdplus_handled)) {
        fputs("The source is protected but libaacs/libbdplus could not handle its protection. Configure authorized decryption data.\n", stderr);
        bd_close(disc);
        return 1;
    }
    if (language_mode) {
        BLURAY_TITLE_INFO *title = bd_get_playlist_info(disc, requested_playlist, 0);
        if (!title || title->playlist != requested_playlist || !title->clip_count ||
            title->clip_count > 10000) {
            fputs("Cannot read requested Blu-ray playlist stream metadata.\n", stderr);
            if (title) bd_free_title_info(title);
            bd_close(disc);
            return 1;
        }
        const int bad = print_stream_languages(title);
        bd_free_title_info(title);
        bd_close(disc);
        return bad;
    }
    uint32_t count = bd_get_titles(disc, 0, 0);
    if (!count || count > 100000) {
        fputs("No Blu-ray playlists found (or an unreasonable title count).\n", stderr);
        bd_close(disc);
        return 1;
    }
    puts("MUKSMATT_BD_NAV_1");
    unsigned emitted = 0;
    /* libbluray may list the same MPLS more than once; retain first angle. */
    unsigned char emitted_playlist[100000] = {0};
    for (uint32_t i = 0; i < count; ++i) {
        BLURAY_TITLE_INFO *title = bd_get_title_info(disc, i, 0);
        if (!title) continue;
        const uint32_t pid = title->playlist;
        const uint64_t duration = title->duration;
        if (pid > 99999 || emitted_playlist[pid] || !duration ||
            duration > UINT64_C(90000) * 60 * 60 * 48 ||
            title->chapter_count > 10000 || title->clip_count == 0 ||
            title->clip_count > 10000) {
            bd_free_title_info(title);
            continue;
        }
        emitted_playlist[pid] = 1;
        printf("P\t%u\t%" PRIu64 "\t%u\n", pid, duration, title->clip_count);
        for (uint32_t ch = 0; ch < title->chapter_count; ++ch) {
            uint64_t start = title->chapters[ch].start;
            uint64_t end = start + title->chapters[ch].duration;
            if (end < start || end > duration || end <= start) continue;
            printf("C\t%u\t%" PRIu64 "\t%" PRIu64 "\n", pid, start, end);
        }
        bd_free_title_info(title);
        ++emitted;
    }
    bd_close(disc);
    if (!emitted) {
        fputs("libbluray enumerated no readable playlists.\n", stderr);
        return 1;
    }
    return 0;
}
