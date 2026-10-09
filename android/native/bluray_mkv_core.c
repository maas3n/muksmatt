/*
 * Android Blu-ray direct one-pass Matroska remux engine.
 *
 * A selected libbluray playlist is fed to FFmpeg via custom AVIO, with no
 * intermediate MKV or decrypted source file. All selected compatible streams
 * are copied EXCEPT Blu-ray LPCM, which is losslessly encoded to FLAC.
 * FFmpeg's MPEG-TS demuxer owns the codec/track identification.
 * DVD input/FFmpeg probe flags are intentionally not applied.
 */
#define _POSIX_C_SOURCE 200809L
#include "bluray_mkv_core.h"
#include <libavcodec/avcodec.h>
#include <libavformat/avformat.h>
#include <libavutil/audio_fifo.h>
#include <libavutil/channel_layout.h>
#include <libavutil/opt.h>
#include <libavutil/mem.h>
#include <libswresample/swresample.h>
#include <errno.h>
#include <inttypes.h>
#include <limits.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>

#define BD_AVIO_CHUNK 6144
#define BD_AVIO_SIZE (BD_AVIO_CHUNK * 32)
#define MAX_BD_STREAMS 256
#define MAX_BD_CHAPTERS 10000

typedef struct {
    BLURAY *bd;
    uint8_t stash[BD_AVIO_CHUNK];
    int stash_pos, stash_len;
    int error;
} BdReader;

typedef struct {
    int fd;
    int64_t pos;
    int seekable;
} BdWriter;

typedef struct {
    int output;
    AVCodecContext *decoder, *encoder;
    SwrContext *swr;
    AVAudioFifo *fifo;
    int64_t next_pts;
} AudioTrack;

static void detail(char *error, size_t size, const char *where, int ret)
{
    char ff[128];
    av_strerror(ret, ff, sizeof(ff));
    snprintf(error, size, "%s: %s", where, ff);
}

static int bd_read_avio(void *opaque, uint8_t *dst, int requested)
{
    BdReader *reader = opaque;
    if (!reader || !dst || requested <= 0) return AVERROR(EINVAL);
    int copied = 0;
    while (copied < requested) {
        if (reader->stash_pos == reader->stash_len) {
            reader->stash_pos = reader->stash_len = 0;
            int got = bd_read(reader->bd, reader->stash, BD_AVIO_CHUNK);
            if (got < 0) { reader->error = got; return copied ? copied : AVERROR(EIO); }
            if (got == 0) return copied ? copied : AVERROR_EOF;
            reader->stash_len = got;
        }
        int available = reader->stash_len - reader->stash_pos;
        int count = requested - copied < available ? requested - copied : available;
        memcpy(dst + copied, reader->stash + reader->stash_pos, (size_t)count);
        copied += count;
        reader->stash_pos += count;
    }
    return copied;
}

static int64_t bd_seek_avio(void *opaque, int64_t offset, int whence)
{
    BdReader *reader = opaque;
    if (!reader) return AVERROR(EINVAL);
    uint64_t total = bd_get_title_size(reader->bd);
    if ((whence & ~AVSEEK_FORCE) == AVSEEK_SIZE)
        return total <= INT64_MAX ? (int64_t)total : AVERROR(EOVERFLOW);
    if (whence & ~(AVSEEK_FORCE | 0x3)) return AVERROR(EINVAL);
    int origin = whence & ~AVSEEK_FORCE;
    uint64_t physical = bd_tell(reader->bd);
    uint64_t logical = physical >= (unsigned)(reader->stash_len - reader->stash_pos)
                         ? physical - (unsigned)(reader->stash_len - reader->stash_pos) : 0;
    int64_t base;
    if (origin == SEEK_SET) base = 0;
    else if (origin == SEEK_CUR) base = logical <= INT64_MAX ? (int64_t)logical : -1;
    else if (origin == SEEK_END) base = total <= INT64_MAX ? (int64_t)total : -1;
    else return AVERROR(EINVAL);
    if (base < 0 || (offset < 0 && offset < -base) ||
        (offset > 0 && offset > INT64_MAX - base)) return AVERROR(EINVAL);
    int64_t position = base + offset;
    if (position < 0 || (uint64_t)position > total) return AVERROR(EINVAL);
    int64_t newpos = bd_seek(reader->bd, (uint64_t)position);
    if (newpos < 0) return AVERROR(EIO);
    reader->stash_pos = reader->stash_len = 0;
    return newpos;
}

static int bd_write_avio(void *opaque, const uint8_t *src, int length)
{
    BdWriter *writer = opaque;
    if (!writer || !src || length < 0) return AVERROR(EINVAL);
    int done = 0;
    while (done < length) {
        ssize_t n = write(writer->fd, src + done, (size_t)(length - done));
        if (n < 0 && errno == EINTR) continue;
        if (n <= 0) return AVERROR(n < 0 ? errno : EIO);
        done += (int)n;
        writer->pos += n;
    }
    return done;
}

static int64_t bd_write_seek(void *opaque, int64_t offset, int whence)
{
    BdWriter *writer = opaque;
    if (!writer || !writer->seekable) return AVERROR(ENOSYS);
    if ((whence & ~AVSEEK_FORCE) == AVSEEK_SIZE) {
        off_t current = lseek(writer->fd, 0, SEEK_CUR);
        if (current < 0) return AVERROR(errno);
        off_t end = lseek(writer->fd, 0, SEEK_END);
        (void)lseek(writer->fd, current, SEEK_SET);
        return end < 0 ? AVERROR(errno) : (int64_t)end;
    }
    if (whence & ~(AVSEEK_FORCE | 0x3)) return AVERROR(EINVAL);
    off_t pos = lseek(writer->fd, (off_t)offset, whence & ~AVSEEK_FORCE);
    if (pos < 0) return AVERROR(errno);
    writer->pos = pos;
    return pos;
}

static BLURAY_TITLE_INFO *select_playlist(BLURAY *bd, int requested)
{
    if (!bd || requested < -1 || requested > 99999) return NULL;
    const BLURAY_DISC_INFO *info = bd_get_disc_info(bd);
    if (!info || (info->aacs_detected && !info->aacs_handled) ||
        (info->bdplus_detected && !info->bdplus_handled)) return NULL;
    uint32_t titles = bd_get_titles(bd, 0, 0);
    if (!titles || titles > 10000) return NULL;
    BLURAY_TITLE_INFO *best = NULL;
    for (uint32_t i = 0; i < titles; ++i) {
        BLURAY_TITLE_INFO *one = bd_get_title_info(bd, i, 0);
        if (!one) continue;
        if (one->playlist > 99999 || !one->duration ||
            one->duration > UINT64_C(90000) * 3600 * 48 ||
            one->clip_count > 10000 || one->chapter_count > MAX_BD_CHAPTERS) {
            bd_free_title_info(one);
            continue;
        }
        int better = requested >= 0 ? (int)one->playlist == requested
                                   : (!best || one->duration > best->duration);
        if (better) {
            if (best) bd_free_title_info(best);
            best = one;
            if (requested >= 0) break;
        } else bd_free_title_info(one);
    }
    if (!best) return NULL;
    if (!bd_select_playlist(bd, best->playlist)) {
        bd_free_title_info(best);
        return NULL;
    }
    return best;
}

static int add_bd_chapters(AVFormatContext *output, const BLURAY_TITLE_INFO *title)
{
    if (!title || title->chapter_count == 0) return 0;
    for (unsigned i = 0; i < title->chapter_count; ++i) {
        const uint64_t start = title->chapters[i].start;
        uint64_t end = i + 1 < title->chapter_count
                         ? title->chapters[i + 1].start : title->duration;
        if (start >= end || end > title->duration) return AVERROR_INVALIDDATA;
        AVChapter *chapter = av_mallocz(sizeof(*chapter));
        if (!chapter) return AVERROR(ENOMEM);
        chapter->id = (int)i;
        chapter->time_base = (AVRational){1, 90000};
        chapter->start = (int64_t)start;
        chapter->end = (int64_t)end;
        AVChapter **entries = av_realloc_array(output->chapters,
                                                output->nb_chapters + 1,
                                                sizeof(*entries));
        if (!entries) { av_free(chapter); return AVERROR(ENOMEM); }
        output->chapters = entries;
        output->chapters[output->nb_chapters++] = chapter;
    }
    return 0;
}

static void audio_free(AudioTrack *a)
{
    if (!a) return;
    avcodec_free_context(&a->decoder);
    avcodec_free_context(&a->encoder);
    swr_free(&a->swr);
    if (a->fifo) av_audio_fifo_free(a->fifo);
    memset(a, 0, sizeof(*a));
}

static int audio_init(AudioTrack *a, const AVStream *source, AVStream *dest)
{
    memset(a, 0, sizeof(*a));
    a->next_pts = AV_NOPTS_VALUE;
    const AVCodec *dec = avcodec_find_decoder(AV_CODEC_ID_PCM_BLURAY);
    const AVCodec *enc = avcodec_find_encoder(AV_CODEC_ID_FLAC);
    if (!dec || !enc) return AVERROR_DECODER_NOT_FOUND;
    a->decoder = avcodec_alloc_context3(dec);
    a->encoder = avcodec_alloc_context3(enc);
    if (!a->decoder || !a->encoder) return AVERROR(ENOMEM);
    int ret = avcodec_parameters_to_context(a->decoder, source->codecpar);
    if (ret < 0 || (ret = avcodec_open2(a->decoder, dec, NULL)) < 0) return ret;
    a->encoder->sample_rate = a->decoder->sample_rate;
    if (a->encoder->sample_rate <= 0) return AVERROR_INVALIDDATA;
    a->encoder->time_base = (AVRational){1, a->encoder->sample_rate};
    a->encoder->sample_fmt = a->decoder->bits_per_raw_sample > 16
                            ? AV_SAMPLE_FMT_S32 : AV_SAMPLE_FMT_S16;
    ret = av_channel_layout_copy(&a->encoder->ch_layout, &a->decoder->ch_layout);
    if (ret < 0) return ret;
    if (!a->encoder->ch_layout.nb_channels) return AVERROR_INVALIDDATA;
    if ((ret = avcodec_open2(a->encoder, enc, NULL)) < 0) return ret;
    ret = avcodec_parameters_from_context(dest->codecpar, a->encoder);
    if (ret < 0) return ret;
    dest->time_base = a->encoder->time_base;
    ret = swr_alloc_set_opts2(&a->swr, &a->encoder->ch_layout,
        a->encoder->sample_fmt, a->encoder->sample_rate, &a->decoder->ch_layout,
        a->decoder->sample_fmt, a->decoder->sample_rate, 0, NULL);
    if (ret < 0 || !a->swr) return ret < 0 ? ret : AVERROR(ENOMEM);
    if ((ret = swr_init(a->swr)) < 0) return ret;
    a->fifo = av_audio_fifo_alloc(a->encoder->sample_fmt,
                                   a->encoder->ch_layout.nb_channels, 4096);
    if (!a->fifo) return AVERROR(ENOMEM);
    a->output = dest->index;
    return 0;
}

static int audio_drain_encoder(AudioTrack *a, AVFormatContext *out)
{
    AVPacket *packet = av_packet_alloc();
    if (!packet) return AVERROR(ENOMEM);
    int ret;
    while ((ret = avcodec_receive_packet(a->encoder, packet)) >= 0) {
        av_packet_rescale_ts(packet, a->encoder->time_base,
                             out->streams[a->output]->time_base);
        packet->stream_index = a->output;
        packet->pos = -1;
        ret = av_interleaved_write_frame(out, packet);
        av_packet_unref(packet);
        if (ret < 0) break;
    }
    av_packet_free(&packet);
    return ret == AVERROR(EAGAIN) || ret == AVERROR_EOF ? 0 : ret;
}

static int audio_emit(AudioTrack *a, AVFormatContext *out, int count)
{
    if (count < 1) return 0;
    AVFrame *frame = av_frame_alloc();
    if (!frame) return AVERROR(ENOMEM);
    frame->nb_samples = count;
    frame->format = a->encoder->sample_fmt;
    frame->sample_rate = a->encoder->sample_rate;
    int ret = av_channel_layout_copy(&frame->ch_layout, &a->encoder->ch_layout);
    if (ret >= 0) ret = av_frame_get_buffer(frame, 0);
    if (ret >= 0 && av_audio_fifo_read(a->fifo, (void **)frame->data, count) != count)
        ret = AVERROR(EIO);
    if (ret >= 0) {
        frame->pts = a->next_pts == AV_NOPTS_VALUE ? 0 : a->next_pts;
        a->next_pts = frame->pts + count;
        ret = avcodec_send_frame(a->encoder, frame);
        if (ret >= 0) ret = audio_drain_encoder(a, out);
    }
    av_frame_free(&frame);
    return ret;
}

static int audio_drain_decoder(AudioTrack *a, AVFormatContext *out,
                               const AVStream *source)
{
    AVFrame *decoded = av_frame_alloc(), *converted = av_frame_alloc();
    if (!decoded || !converted) {
        av_frame_free(&decoded); av_frame_free(&converted);
        return AVERROR(ENOMEM);
    }
    int ret;
    while ((ret = avcodec_receive_frame(a->decoder, decoded)) >= 0) {
        if (a->next_pts == AV_NOPTS_VALUE) {
            a->next_pts = decoded->best_effort_timestamp == AV_NOPTS_VALUE ? 0 :
                av_rescale_q(decoded->best_effort_timestamp, source->time_base,
                             a->encoder->time_base);
        }
        converted->format = a->encoder->sample_fmt;
        converted->sample_rate = a->encoder->sample_rate;
        converted->nb_samples = decoded->nb_samples + 32;
        ret = av_channel_layout_copy(&converted->ch_layout, &a->encoder->ch_layout);
        if (ret >= 0) ret = av_frame_get_buffer(converted, 0);
        if (ret >= 0) ret = swr_convert(a->swr, converted->data,
                             converted->nb_samples,
                             (const uint8_t **)decoded->extended_data,
                             decoded->nb_samples);
        if (ret >= 0) {
            if (av_audio_fifo_realloc(a->fifo, av_audio_fifo_size(a->fifo) + ret) < 0 ||
                av_audio_fifo_write(a->fifo, (void **)converted->data, ret) != ret)
                ret = AVERROR(ENOMEM);
            else ret = 0;
        }
        av_frame_unref(decoded);
        av_frame_unref(converted);
        if (ret < 0) break;
        int frame_size = a->encoder->frame_size > 0 ? a->encoder->frame_size : 4096;
        while (av_audio_fifo_size(a->fifo) >= frame_size) {
            ret = audio_emit(a, out, frame_size);
            if (ret < 0) break;
        }
        if (ret < 0) break;
    }
    av_frame_free(&decoded); av_frame_free(&converted);
    return ret == AVERROR(EAGAIN) || ret == AVERROR_EOF ? 0 : ret;
}

static int audio_packet(AudioTrack *a, AVFormatContext *out,
                         const AVStream *source, const AVPacket *packet)
{
    int ret = avcodec_send_packet(a->decoder, packet);
    if (ret < 0) return ret;
    return audio_drain_decoder(a, out, source);
}

static int audio_finish(AudioTrack *a, AVFormatContext *out,
                        const AVStream *source)
{
    int ret = avcodec_send_packet(a->decoder, NULL);
    if (ret >= 0) ret = audio_drain_decoder(a, out, source);
    if (ret < 0) return ret;
    /* Sample rate is unchanged; the conversion has no time-stretching. */
    int remain = av_audio_fifo_size(a->fifo);
    if (remain) {
        ret = audio_emit(a, out, remain);
        if (ret < 0) return ret;
    }
    if ((ret = avcodec_send_frame(a->encoder, NULL)) < 0) return ret;
    return audio_drain_encoder(a, out);
}

int muksmatt_bd_mkv(BLURAY *bd, int requested_playlist, int output_fd,
                     const int *selection, int selection_count,
                     int include_chapters, char *error, size_t error_size)
{
    int ret = AVERROR_INVALIDDATA;
    BLURAY_TITLE_INFO *title = NULL;
    AVFormatContext *input = NULL, *out = NULL;
    AVIOContext *input_io = NULL, *output_io = NULL;
    AudioTrack audio[MAX_BD_STREAMS] = {0};
    int mapping[MAX_BD_STREAMS];
    AVPacket *packet = NULL;
    BdReader reader = {.bd = bd};
    BdWriter writer = {.fd = -1, .pos = 0, .seekable = 0};
    int out_started = 0, selected_count = 0;
    for (int i = 0; i < MAX_BD_STREAMS; ++i) mapping[i] = -1;
    if (!bd || output_fd < 0 || !error || !error_size ||
        requested_playlist < -1 || requested_playlist > 99999 ||
        selection_count < 0 || selection_count > MAX_BD_STREAMS ||
        (selection_count > 0 && !selection)) {
        if (error && error_size) snprintf(error, error_size, "Invalid Blu-ray remux arguments");
        return AVERROR(EINVAL);
    }
    snprintf(error, error_size, "Blu-ray remux failed");
    title = select_playlist(bd, requested_playlist);
    if (!title) {
        snprintf(error, error_size, "Cannot select Blu-ray playlist or protection is not handled");
        goto cleanup;
    }
    writer.fd = dup(output_fd);
    if (writer.fd < 0) {
        snprintf(error, error_size, "Cannot duplicate SAF output descriptor");
        goto cleanup;
    }
    off_t first = lseek(writer.fd, 0, SEEK_CUR);
    writer.seekable = first >= 0;
    if (writer.seekable && lseek(writer.fd, 0, SEEK_SET) < 0) {
        snprintf(error, error_size, "Cannot seek SAF output descriptor");
        goto cleanup;
    }
    uint8_t *in_buf = av_malloc(BD_AVIO_SIZE);
    if (!in_buf) { ret = AVERROR(ENOMEM); goto cleanup; }
    input_io = avio_alloc_context(in_buf, BD_AVIO_SIZE, 0, &reader,
                                  bd_read_avio, NULL, bd_seek_avio);
    if (!input_io) { av_free(in_buf); ret = AVERROR(ENOMEM); goto cleanup; }
    input = avformat_alloc_context();
    if (!input) { ret = AVERROR(ENOMEM); goto cleanup; }
    input->pb = input_io;
    input->flags |= AVFMT_FLAG_CUSTOM_IO;
    const AVInputFormat *mpegts = av_find_input_format("mpegts");
    if (!mpegts) { ret = AVERROR_DEMUXER_NOT_FOUND; goto cleanup; }
    ret = avformat_open_input(&input, NULL, mpegts, NULL);
    if (ret < 0) { detail(error, error_size, "Reading Blu-ray playlist", ret); goto cleanup; }
    ret = avformat_find_stream_info(input, NULL);
    if (ret < 0) { detail(error, error_size, "Blu-ray stream discovery", ret); goto cleanup; }
    if (input->nb_streams == 0 || input->nb_streams > MAX_BD_STREAMS) {
        ret = AVERROR_INVALIDDATA;
        snprintf(error, error_size, "Blu-ray has unsupported stream count");
        goto cleanup;
    }
    if (selection_count > 0) {
        for (int i = 0; i < selection_count; ++i) {
            if (selection[i] < 0 || (unsigned)selection[i] >= input->nb_streams) {
                ret = AVERROR(EINVAL);
                snprintf(error, error_size, "Selected Blu-ray track index is invalid");
                goto cleanup;
            }
            for (int j = 0; j < i; ++j) if (selection[i] == selection[j]) {
                ret = AVERROR(EINVAL);
                snprintf(error, error_size, "Duplicate Blu-ray track selection");
                goto cleanup;
            }
        }
    }
    ret = avformat_alloc_output_context2(&out, NULL, "matroska", NULL);
    if (ret < 0 || !out) goto cleanup;
    for (unsigned i = 0; i < input->nb_streams; ++i) {
        AVStream *src = input->streams[i];
        enum AVMediaType kind = src->codecpar->codec_type;
        if (kind != AVMEDIA_TYPE_VIDEO && kind != AVMEDIA_TYPE_AUDIO &&
            kind != AVMEDIA_TYPE_SUBTITLE) continue;
        int enabled = selection_count == 0;
        for (int j = 0; j < selection_count; ++j) if (selection[j] == (int)i) enabled = 1;
        if (!enabled) continue;
        AVStream *dest = avformat_new_stream(out, NULL);
        if (!dest) { ret = AVERROR(ENOMEM); goto cleanup; }
        mapping[i] = dest->index;
        dest->disposition = src->disposition;
        av_dict_copy(&dest->metadata, src->metadata, 0);
        if (src->codecpar->codec_id == AV_CODEC_ID_PCM_BLURAY) {
            ret = audio_init(&audio[i], src, dest);
            if (ret < 0) {
                detail(error, error_size, "Configuring mandatory LPCM-to-FLAC conversion", ret);
                goto cleanup;
            }
        } else {
            ret = avcodec_parameters_copy(dest->codecpar, src->codecpar);
            if (ret < 0) goto cleanup;
            dest->codecpar->codec_tag = 0;
            dest->time_base = src->time_base;
        }
        ++selected_count;
    }
    if (!selected_count) {
        ret = AVERROR(EINVAL);
        snprintf(error, error_size, "No compatible selected Blu-ray streams");
        goto cleanup;
    }
    if (include_chapters) {
        ret = add_bd_chapters(out, title);
        if (ret < 0) { detail(error, error_size, "Invalid Blu-ray chapters", ret); goto cleanup; }
    }
    uint8_t *out_buf = av_malloc(64 * 1024);
    if (!out_buf) { ret = AVERROR(ENOMEM); goto cleanup; }
    output_io = avio_alloc_context(out_buf, 64 * 1024, 1, &writer, NULL,
                                   bd_write_avio, bd_write_seek);
    if (!output_io) { av_free(out_buf); ret = AVERROR(ENOMEM); goto cleanup; }
    output_io->seekable = writer.seekable ? AVIO_SEEKABLE_NORMAL : 0;
    out->pb = output_io;
    out->flags |= AVFMT_FLAG_CUSTOM_IO;
    ret = avformat_write_header(out, NULL);
    if (ret < 0) { detail(error, error_size, "Cannot create Blu-ray Matroska header", ret); goto cleanup; }
    out_started = 1;
    packet = av_packet_alloc();
    if (!packet) { ret = AVERROR(ENOMEM); goto cleanup; }
    while ((ret = av_read_frame(input, packet)) >= 0) {
        int index = packet->stream_index;
        if (index >= 0 && index < MAX_BD_STREAMS && mapping[index] >= 0) {
            AVStream *src = input->streams[index], *dst = out->streams[mapping[index]];
            if (src->codecpar->codec_id == AV_CODEC_ID_PCM_BLURAY) {
                ret = audio_packet(&audio[index], out, src, packet);
            } else {
                av_packet_rescale_ts(packet, src->time_base, dst->time_base);
                packet->stream_index = mapping[index];
                packet->pos = -1;
                ret = av_interleaved_write_frame(out, packet);
            }
        } else ret = 0;
        av_packet_unref(packet);
        if (ret < 0) { detail(error, error_size, "Blu-ray packet copy/FLAC encode", ret); goto cleanup; }
        if (reader.error < 0) {
            ret = AVERROR(EIO);
            snprintf(error, error_size, "Blu-ray source read failed");
            goto cleanup;
        }
    }
    if (ret != AVERROR_EOF) {
        detail(error, error_size, "Blu-ray transport stream read", ret);
        goto cleanup;
    }
    for (unsigned i = 0; i < input->nb_streams; ++i) {
        if (audio[i].decoder) {
            ret = audio_finish(&audio[i], out, input->streams[i]);
            if (ret < 0) { detail(error, error_size, "Finishing lossless FLAC track", ret); goto cleanup; }
        }
    }
    ret = av_write_trailer(out);
    if (ret >= 0) {
        avio_flush(output_io);
        if (output_io->error < 0) ret = output_io->error;
    }
    if (ret < 0) detail(error, error_size, "Finalizing Blu-ray MKV", ret);
    else error[0] = '\0';

cleanup:
    av_packet_free(&packet);
    if (out_started && ret < 0) avio_flush(output_io);
    for (int i = 0; i < MAX_BD_STREAMS; ++i) audio_free(&audio[i]);
    if (out) avformat_free_context(out);
    if (output_io) { av_freep(&output_io->buffer); avio_context_free(&output_io); }
    if (input) {
        if (input->iformat) avformat_close_input(&input);
        else avformat_free_context(input);
    }
    if (input_io) { av_freep(&input_io->buffer); avio_context_free(&input_io); }
    if (title) bd_free_title_info(title);
    if (writer.fd >= 0) close(writer.fd);
    if (ret < 0 && error[0] == '\0') detail(error, error_size, "Blu-ray remux failed", ret);
    return ret;
}
