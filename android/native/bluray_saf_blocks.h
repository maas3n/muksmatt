/* Seekable Android SAF ISO sector reader (read-only, no DVD engine).
 * libbluray bd_open_stream() uses 2048-byte disc logical blocks.
 */
#ifndef MUKSMATT_BLURAY_SAF_BLOCKS_H
#define MUKSMATT_BLURAY_SAF_BLOCKS_H
#include <stdint.h>
typedef struct {
    int fd; /* owned duplicate, -1 when closed */
    int64_t blocks;
} BluraySafBlocks;
int bluray_saf_blocks_open(BluraySafBlocks *out, int borrowed_fd);
int bluray_saf_read_blocks(void *handle, void *buffer, int lba, int count);
void bluray_saf_blocks_close(BluraySafBlocks *reader);
#endif
