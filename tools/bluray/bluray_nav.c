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
    fputs("Usage: muksmatt-bluray-nav SOURCE (Blu-ray folder, ISO, or optical device)\n", stderr);
    return 2;
}

int main(int argc, char **argv) {
    if (argc != 2 || !argv[1] || !argv[1][0]) return usage();
    if (!strcmp(argv[1], "--version")) {
        puts("MUKSMATT_BD_NAV_1");
        return 0;
    }
    BLURAY *disc = bd_open(argv[1], NULL);
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
