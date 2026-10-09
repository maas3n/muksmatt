/* Android Blu-ray native ISO inspection via libbluray and a seekable SAF FD.
 *
 * This is intentionally a separate native .so from the stable DVD JNI engine.
 * It DOES NOT create an MKV yet: it verifies real libbluray disc/playlist access
 * and a bounded read of decrypted or unprotected M2TS payload from SAF.
 * Neither AACS/BD+ keys nor DVD APIs are present in this implementation.
 */
#define _POSIX_C_SOURCE 200809L
#include <jni.h>
#include <libbluray/bluray.h>
#include "bluray_saf_blocks.h"
#include "bluray_mkv_core.h"
#include <limits.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

static void io_error(JNIEnv *env, const char *msg)
{
    jclass cls = (*env)->FindClass(env, "java/io/IOException");
    if (cls) (*env)->ThrowNew(env, cls, msg);
}

JNIEXPORT jstring JNICALL
Java_io_github_maas3n_mattmux_BlurayNativeIsoBridge_nativeInspectIso(
    JNIEnv *env, jobject self, jint input_fd, jint requested_playlist, jint sample_bytes)
{
    (void)self;
    if (requested_playlist < -1 || requested_playlist > 99999 ||
        sample_bytes < 0 || sample_bytes > 6144) {
        io_error(env, "Invalid Blu-ray playlist or sample size");
        return NULL;
    }

    BluraySafBlocks blocks = {.fd = -1, .blocks = 0};
    BLURAY *bd = NULL;
    char *result = NULL;
    jstring response = NULL;
    const char *failure = NULL;
    if (bluray_saf_blocks_open(&blocks, input_fd) < 0) {
        io_error(env, "Blu-ray ISO must be a complete, seekable UDF image from a SAF provider");
        return NULL;
    }
    bd = bd_init();
    if (!bd || !bd_open_stream(bd, &blocks, bluray_saf_read_blocks)) {
        failure = "libbluray could not open the Blu-ray ISO (unsupported or invalid UDF image)";
        goto cleanup;
    }

    const BLURAY_DISC_INFO *info = bd_get_disc_info(bd);
    if (!info || (info->aacs_detected && !info->aacs_handled) ||
        (info->bdplus_detected && !info->bdplus_handled)) {
        failure = "Blu-ray protection is not handled: authorized local AACS/BD+ support is required";
        goto cleanup;
    }
    uint32_t title_count = bd_get_titles(bd, 0, 0);
    if (!title_count || title_count > 10000) {
        failure = "No readable Blu-ray titles were discovered";
        goto cleanup;
    }

    result = calloc(1, 4096);
    if (!result) {
        failure = "Out of memory";
        goto cleanup;
    }
    size_t used = (size_t)snprintf(result, 4096, "MUKSMATT_ANDROID_BD_1\n");
    uint64_t longest = 0;
    uint32_t selected = UINT32_MAX;
    uint32_t found = 0;
    for (uint32_t i = 0; i < title_count; ++i) {
        BLURAY_TITLE_INFO *title = bd_get_title_info(bd, i, 0);
        if (!title) continue;
        if (title->playlist <= 99999 && title->duration > 0 &&
            title->duration <= UINT64_C(90000) * 3600 * 48 &&
            title->clip_count <= 10000 && title->chapter_count <= 10000) {
            if (requested_playlist == (int)title->playlist ||
                (requested_playlist < 0 && title->duration > longest)) {
                longest = title->duration;
                selected = title->playlist;
            }
            ++found;
        }
        bd_free_title_info(title);
    }
    if (!found || selected == UINT32_MAX) {
        failure = "Requested Blu-ray playlist was not found in the ISO";
        goto cleanup;
    }
    if (!bd_select_playlist(bd, selected)) {
        failure = "libbluray could not select the requested Blu-ray playlist";
        goto cleanup;
    }
    unsigned bytes_read = 0;
    if (sample_bytes) {
        unsigned char sample[6144];
        int got = bd_read(bd, sample, sample_bytes);
        if (got <= 0) {
            failure = "libbluray could not read Blu-ray video packets (check disc access/decryption)";
            goto cleanup;
        }
        bytes_read = (unsigned)got;
    }
    if (used >= 4096 ||
        snprintf(result + used, 4096 - used, "P\t%u\t%llu\nR\t%u\n",
            selected, (unsigned long long)longest, bytes_read) >= (int)(4096 - used)) {
        failure = "Blu-ray playlist result is oversized";
        goto cleanup;
    }
    response = (*env)->NewStringUTF(env, result);

cleanup:
    if (bd) bd_close(bd);
    bluray_saf_blocks_close(&blocks);
    free(result);
    if (failure && !(*env)->ExceptionCheck(env)) io_error(env, failure);
    return response;
}


/* Write a selected Blu-ray playlist straight to a SAF MKV output descriptor.
 * Both SAF descriptors are borrowed and remain owned by Kotlin.
 * Returns NULL on success, descriptive error on failure.
 */
JNIEXPORT jstring JNICALL
Java_io_github_maas3n_mattmux_BlurayNativeIsoBridge_nativeRemuxIso(
    JNIEnv *env, jobject self, jint iso_fd, jint output_fd, jint playlist,
    jintArray selected_streams, jboolean chapters)
{
    (void)self;
    if (iso_fd < 0 || output_fd < 0 || playlist < -1 || playlist > 99999) {
        return (*env)->NewStringUTF(env, "Invalid Blu-ray ISO remux descriptors or playlist");
    }
    int count = selected_streams ? (*env)->GetArrayLength(env, selected_streams) : 0;
    if (count < 0 || count > 256)
        return (*env)->NewStringUTF(env, "Invalid Blu-ray stream selection");
    jint *indices = selected_streams ? (*env)->GetIntArrayElements(env, selected_streams, NULL) : NULL;
    if (selected_streams && !indices) return NULL;
    BluraySafBlocks blocks = {.fd = -1};
    BLURAY *bd = NULL;
    char error[512] = "Could not open Blu-ray ISO";
    int ret = -1;
    if (bluray_saf_blocks_open(&blocks, iso_fd) < 0) {
        snprintf(error, sizeof(error), "Blu-ray ISO must be a complete seekable SAF file");
    } else {
        bd = bd_init();
        if (!bd || !bd_open_stream(bd, &blocks, bluray_saf_read_blocks))
            snprintf(error, sizeof(error), "libbluray cannot open the selected Blu-ray ISO");
        else ret = muksmatt_bd_mkv(bd, playlist, output_fd, (const int *)indices,
                                    count, chapters ? 1 : 0, error, sizeof(error));
    }
    if (bd) bd_close(bd);
    bluray_saf_blocks_close(&blocks);
    if (indices) (*env)->ReleaseIntArrayElements(env, selected_streams, indices, JNI_ABORT);
    return ret < 0 ? (*env)->NewStringUTF(env, error) : NULL;
}


JNIEXPORT jstring JNICALL
Java_io_github_maas3n_mattmux_BlurayNativeIsoBridge_nativeProbeStreamsIso(
    JNIEnv *env, jobject self, jint iso_fd, jint playlist)
{
    (void)self;
    if (iso_fd < 0 || playlist < -1 || playlist > 99999) {
        io_error(env, "Invalid Blu-ray ISO stream probe arguments");
        return NULL;
    }
    BluraySafBlocks blocks = {.fd = -1};
    BLURAY *bd = NULL;
    char report[32768] = {0};
    char error[512] = "Cannot open Blu-ray ISO";
    int ret = -1;
    if (bluray_saf_blocks_open(&blocks, iso_fd) >= 0) {
        bd = bd_init();
        if (bd && bd_open_stream(bd, &blocks, bluray_saf_read_blocks))
            ret = muksmatt_bd_tracks(bd, playlist, report, sizeof(report),
                                     error, sizeof(error));
    }
    if (bd) bd_close(bd);
    bluray_saf_blocks_close(&blocks);
    if (ret < 0) {
        io_error(env, error);
        return NULL;
    }
    return (*env)->NewStringUTF(env, report);
}
