#ifndef MUKSMATT_BLURAY_MKV_CORE_H
#define MUKSMATT_BLURAY_MKV_CORE_H
#include <libbluray/bluray.h>
#include <stddef.h>
/*
 * One-pass Blu-ray playlist -> Matroska. Selection indices refer to MPEG-TS
 * AVStream indices as enumerated by avformat_find_stream_info(). NULL selects
 * all supported AV streams. PCM_BLURAY is always transcoded losslessly to FLAC.
 * The owned BLURAY pointer remains with the caller and is not closed here.
 */
int muksmatt_bd_mkv(BLURAY *bd, int requested_playlist, int output_fd,
                     const int *selection, int selection_count,
                     int include_chapters, char *error, size_t error_size);
/* Discover MPEG-TS stream indices for subsequent selection. Emits a bounded
 * versioned TSV with one S/index/type/codec row for each eligible track. */
int muksmatt_bd_tracks(BLURAY *bd, int requested_playlist, char *report,
                        size_t report_size, char *error, size_t error_size);
#endif
