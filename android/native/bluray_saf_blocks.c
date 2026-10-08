#define _POSIX_C_SOURCE 200809L
#include "bluray_saf_blocks.h"
#include <errno.h>
#include <limits.h>
#include <stddef.h>
#include <stdint.h>
#include <string.h>
#include <sys/stat.h>
#include <sys/types.h>
#include <unistd.h>

#define BD_BLOCK_BYTES 2048

int bluray_saf_blocks_open(BluraySafBlocks *out, int borrowed_fd)
{
    if (!out) return -1;
    out->fd = -1;
    out->blocks = 0;
    if (borrowed_fd < 0) return -1;
    struct stat st;
    unsigned char probe;
    if (fstat(borrowed_fd, &st) || st.st_size < 257LL * BD_BLOCK_BYTES ||
        st.st_size % BD_BLOCK_BYTES || st.st_size / BD_BLOCK_BYTES > INT_MAX ||
        pread(borrowed_fd, &probe, 1, 0) != 1) {
        return -1; /* Pipes, empty/truncated images and unknown-size FDs fail closed. */
    }
    int owned = dup(borrowed_fd);
    if (owned < 0) return -1;
    out->fd = owned;
    out->blocks = st.st_size / BD_BLOCK_BYTES;
    return 0;
}

int bluray_saf_read_blocks(void *handle, void *buffer, int lba, int count)
{
    BluraySafBlocks *reader = handle;
    if (!reader || reader->fd < 0 || !buffer || lba < 0 || count <= 0 ||
        (int64_t)lba > reader->blocks ||
        (int64_t)count > reader->blocks - lba ||
        count > INT_MAX / BD_BLOCK_BYTES) {
        return -1;
    }
    const size_t size = (size_t)count * BD_BLOCK_BYTES;
    size_t done = 0;
    while (done < size) {
        const off_t offset = (off_t)lba * BD_BLOCK_BYTES + (off_t)done;
        ssize_t got = pread(reader->fd, (unsigned char *)buffer + done, size - done, offset);
        if (got < 0 && errno == EINTR) continue;
        if (got <= 0) return -1;
        done += (size_t)got;
    }
    return count;
}

void bluray_saf_blocks_close(BluraySafBlocks *reader)
{
    if (!reader) return;
    if (reader->fd >= 0) close(reader->fd);
    reader->fd = -1;
    reader->blocks = 0;
}
