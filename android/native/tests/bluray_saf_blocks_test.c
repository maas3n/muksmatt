/* Portable unit tests for the Android seekable ISO block callback.
 * Test with: cc -std=c11 -Wall -Wextra -Werror -O2
 *   android/native/bluray_saf_blocks.c android/native/tests/bluray_saf_blocks_test.c
 */
#define _POSIX_C_SOURCE 200809L
#include "../bluray_saf_blocks.h"
#include <assert.h>
#include <errno.h>
#include <fcntl.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>

int main(void)
{
    char name[] = "/tmp/muksmatt-bluray-iso-XXXXXX";
    int fd = mkstemp(name);
    assert(fd >= 0);
    assert(unlink(name) == 0);
    const int64_t sectors = 300;
    assert(ftruncate(fd, sectors * 2048) == 0);
    unsigned char data[4096];
    memset(data, 0x5a, sizeof(data));
    assert(pwrite(fd, data, sizeof(data), 257LL * 2048) == sizeof(data));
    BluraySafBlocks reader = {.fd = -1};
    assert(bluray_saf_blocks_open(&reader, fd) == 0);
    assert(reader.blocks == sectors);
    memset(data, 0, sizeof(data));
    assert(bluray_saf_read_blocks(&reader, data, 257, 2) == 2);
    for (size_t i = 0; i < sizeof(data); ++i) assert(data[i] == 0x5a);
    assert(bluray_saf_read_blocks(&reader, data, 299, 2) < 0);
    assert(bluray_saf_read_blocks(&reader, data, -1, 1) < 0);
    assert(bluray_saf_read_blocks(&reader, data, 0, 0) < 0);
    assert(bluray_saf_read_blocks(&reader, NULL, 0, 1) < 0);
    assert(bluray_saf_read_blocks(NULL, data, 0, 1) < 0);
    assert(bluray_saf_read_blocks(&reader, data, 0, INT32_MAX) < 0);
    bluray_saf_blocks_close(&reader);
    assert(bluray_saf_read_blocks(&reader, data, 0, 1) < 0);
    assert(pread(fd, data, 1, 257LL * 2048) == 1); /* user FD ownership kept */
    assert(ftruncate(fd, 256LL * 2048) == 0);
    assert(bluray_saf_blocks_open(&reader, fd) < 0);
    close(fd);
    int pipefd[2];
    assert(pipe(pipefd) == 0);
    assert(bluray_saf_blocks_open(&reader, pipefd[0]) < 0);
    close(pipefd[0]);
    close(pipefd[1]);
    puts("Blu-ray SAF seekable block reader: all tests passed");
    return 0;
}
