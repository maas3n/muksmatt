/*
 * Linux host integration harness for Android's SAME Blu-ray remux core.
 * Never ship to the Android APK. Synthetic media generated in CI.
 */
#define _POSIX_C_SOURCE 200809L
#include "../bluray_mkv_core.h"
#include <libbluray/bluray.h>
#include <errno.h>
#include <fcntl.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>

int main(int argc, char **argv)
{
    if (argc != 4) {
        fprintf(stderr, "usage: %s <BDMV-root-or-ISO> <out.mkv> <all|video|audio>\n", argv[0]);
        return 2;
    }
    BLURAY *bd = bd_open(argv[1], NULL);
    if (!bd) { fprintf(stderr, "bd_open failed: %s\n", argv[1]); return 3; }
    char catalog[32768], error[512];
    int ret = muksmatt_bd_tracks(bd, -1, catalog, sizeof(catalog), error, sizeof(error));
    if (ret < 0) {
        fprintf(stderr, "bd_tracks failed: %s\n", error);
        bd_close(bd);
        return 4;
    }
    printf("%s", catalog);
    int streams[256];
    int chosen = 0;
    if (strcmp(argv[3], "all")) {
        char *copy = strdup(catalog);
        if (!copy) { bd_close(bd); return 5; }
        char *save = NULL;
        for (char *line = strtok_r(copy, "\n", &save); line; line = strtok_r(NULL, "\n", &save)) {
            int index = -1;
            char kind[32], codec[96];
            if (sscanf(line, "S\t%d\t%31[^\t]\t%95s", &index, kind, codec) != 3) continue;
            if (index >= 0 && index < 256 &&
                ((!strcmp(argv[3], "video") && !strcmp(kind, "video")) ||
                 (!strcmp(argv[3], "audio") && !strcmp(kind, "audio"))))
                streams[chosen++] = index;
        }
        free(copy);
        if (!chosen) { fprintf(stderr, "Requested stream type missing\n"); bd_close(bd); return 5; }
    }
    int fd = open(argv[2], O_WRONLY | O_CREAT | O_TRUNC, 0600);
    if (fd < 0) { perror("open output"); bd_close(bd); return 6; }
    ret = muksmatt_bd_mkv(bd, -1, fd, chosen ? streams : NULL,
                           chosen, 1, error, sizeof(error));
    if (close(fd)) { perror("close output"); ret = -1; }
    bd_close(bd);
    if (ret < 0) { fprintf(stderr, "bd_mkv failed: %s\n", error); return 7; }
    return 0;
}
