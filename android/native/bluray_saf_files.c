/*
 * Read-only libbluray bd_open_files() callbacks backed by Android SAF.
 * All file/directory lookups are confined to BluraySafPathIndex's user-granted
 * tree. The DVD engine and its libudfread runtime are not used here.
 */
#define _POSIX_C_SOURCE 200809L
#include <jni.h>
#include <libbluray/bluray.h>
#include <libbluray/filesystem.h>
#include <errno.h>
#include <limits.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/stat.h>
#include <unistd.h>

typedef struct {
    JNIEnv *env;
    jobject provider;
    jmethodID open_fd;
    jmethodID list_names;
    int callback_error;
} SafTree;

typedef struct {
    BD_DIR_H api;
    char **names;
    int count;
    int next;
} SafDir;

typedef struct {
    BD_FILE_H api;
    int fd;
    int64_t pos;
    int64_t size;
} SafFile;

static void saf_dir_close(BD_DIR_H *base)
{
    SafDir *dir = (SafDir *)base;
    if (!dir) return;
    for (int i = 0; i < dir->count; ++i) free(dir->names[i]);
    free(dir->names);
    free(dir);
}
static int saf_dir_read(BD_DIR_H *base, BD_DIRENT *entry)
{
    SafDir *dir = (SafDir *)base;
    if (!dir || !entry) return -1;
    if (dir->next == dir->count) return 1;
    memset(entry, 0, sizeof(*entry));
    strncpy(entry->d_name, dir->names[dir->next++], sizeof(entry->d_name) - 1);
    return 0;
}
static BD_DIR_H *saf_open_dir(void *opaque, const char *path)
{
    SafTree *tree = (SafTree *)opaque;
    JNIEnv *env = tree->env;
    if (!path || strlen(path) > 1024) return NULL;
    jstring name = (*env)->NewStringUTF(env, path);
    if (!name) return NULL;
    jobjectArray array = (jobjectArray)(*env)->CallObjectMethod(
        env, tree->provider, tree->list_names, name);
    (*env)->DeleteLocalRef(env, name);
    if ((*env)->ExceptionCheck(env)) {
        (*env)->ExceptionClear(env);
        tree->callback_error = 1;
        return NULL;
    }
    if (!array) return NULL; /* absent directory, not a fatal error */
    jsize count = (*env)->GetArrayLength(env, array);
    if (count < 0 || count > 10000) {
        (*env)->DeleteLocalRef(env, array);
        tree->callback_error = 1;
        return NULL;
    }
    SafDir *dir = calloc(1, sizeof(*dir));
    if (!dir) { (*env)->DeleteLocalRef(env, array); return NULL; }
    dir->names = calloc((size_t)count + 1, sizeof(char *));
    if (!dir->names) {
        (*env)->DeleteLocalRef(env, array);
        free(dir);
        return NULL;
    }
    dir->count = (int)count;
    dir->api.close = saf_dir_close;
    dir->api.read = saf_dir_read;
    for (jsize i = 0; i < count; ++i) {
        jstring str = (jstring)(*env)->GetObjectArrayElement(env, array, i);
        if (!str || (*env)->ExceptionCheck(env)) {
            tree->callback_error = 1;
            if ((*env)->ExceptionCheck(env)) (*env)->ExceptionClear(env);
            if (str) (*env)->DeleteLocalRef(env, str);
            saf_dir_close(&dir->api);
            (*env)->DeleteLocalRef(env, array);
            return NULL;
        }
        const char *raw = (*env)->GetStringUTFChars(env, str, NULL);
        size_t len = raw ? strlen(raw) : 0;
        if (len < 1 || len >= sizeof(((BD_DIRENT *)0)->d_name) ||
            strchr(raw, '/') || strchr(raw, '\\') ||
            !strcmp(raw, ".") || !strcmp(raw, "..")) {
            tree->callback_error = 1;
        } else {
            dir->names[i] = strdup(raw);
            if (!dir->names[i]) tree->callback_error = 1;
        }
        if (raw) (*env)->ReleaseStringUTFChars(env, str, raw);
        (*env)->DeleteLocalRef(env, str);
        if (tree->callback_error) {
            if ((*env)->ExceptionCheck(env)) (*env)->ExceptionClear(env);
            saf_dir_close(&dir->api);
            (*env)->DeleteLocalRef(env, array);
            return NULL;
        }
    }
    (*env)->DeleteLocalRef(env, array);
    return &dir->api;
}

static void saf_file_close(BD_FILE_H *base)
{
    SafFile *file = (SafFile *)base;
    if (!file) return;
    if (file->fd >= 0) close(file->fd);
    free(file);
}
static int64_t saf_file_seek(BD_FILE_H *base, int64_t offset, int32_t origin)
{
    SafFile *file = (SafFile *)base;
    if (!file || file->fd < 0) return -1;
    int64_t start;
    if (origin == SEEK_SET) start = 0;
    else if (origin == SEEK_CUR) start = file->pos;
    else if (origin == SEEK_END) start = file->size;
    else return -1;
    if (offset < -start || offset > INT64_MAX - start) return -1;
    int64_t result = start + offset;
    if (result < 0 || result > file->size) return -1;
    file->pos = result;
    return result;
}
static int64_t saf_file_tell(BD_FILE_H *base)
{
    SafFile *f = (SafFile *)base;
    return f ? f->pos : -1;
}
static int saf_file_eof(BD_FILE_H *base)
{
    SafFile *f = (SafFile *)base;
    return !f || f->pos >= f->size;
}
static int64_t saf_file_read(BD_FILE_H *base, uint8_t *data, int64_t size)
{
    SafFile *f = (SafFile *)base;
    if (!f || !data || size < 0 || f->fd < 0) return -1;
    if (!size || f->pos >= f->size) return 0;
    /* Preempt huge allocations/reads; libbluray may call again to continue. */
    if (size > 4 * 1024 * 1024) size = 4 * 1024 * 1024;
    if (size > f->size - f->pos) size = f->size - f->pos;
    ssize_t n;
    do { n = pread(f->fd, data, (size_t)size, (off_t)f->pos); }
    while (n < 0 && errno == EINTR);
    if (n < 0) return -1;
    f->pos += n;
    return n;
}
static int64_t saf_file_write(BD_FILE_H *base, const uint8_t *data, int64_t size)
{
    (void)base; (void)data; (void)size;
    return -1; /* Blu-ray providers are strictly read-only */
}
static BD_FILE_H *saf_open_file(void *opaque, const char *path)
{
    SafTree *tree = (SafTree *)opaque;
    JNIEnv *env = tree->env;
    if (!path || strlen(path) > 1024) return NULL;
    jstring name = (*env)->NewStringUTF(env, path);
    if (!name) return NULL;
    jint fd = (*env)->CallIntMethod(env, tree->provider, tree->open_fd, name);
    (*env)->DeleteLocalRef(env, name);
    if ((*env)->ExceptionCheck(env)) {
        (*env)->ExceptionClear(env);
        tree->callback_error = 1;
        return NULL;
    }
    if (fd < 0) return NULL;
    struct stat st;
    unsigned char probe;
    if (fstat(fd, &st) || st.st_size <= 0 ||
        pread(fd, &probe, 1, 0) != 1) {
        close(fd);
        return NULL;
    }
    SafFile *file = calloc(1, sizeof(*file));
    if (!file) { close(fd); return NULL; }
    file->fd = fd;
    file->size = st.st_size;
    file->api.close = saf_file_close;
    file->api.seek = saf_file_seek;
    file->api.tell = saf_file_tell;
    file->api.eof = saf_file_eof;
    file->api.read = saf_file_read;
    file->api.write = saf_file_write;
    return &file->api;
}

static void io_error(JNIEnv *env, const char *message)
{
    if ((*env)->ExceptionCheck(env)) return;
    jclass cls = (*env)->FindClass(env, "java/io/IOException");
    if (cls) (*env)->ThrowNew(env, cls, message);
}

JNIEXPORT jstring JNICALL
Java_io_github_maas3n_mattmux_BluraySafTreeBridge_nativeInspectTree(
    JNIEnv *env, jobject self, jobject provider, jint requested_playlist, jint sample_bytes)
{
    (void)self;
    if (!provider || requested_playlist < -1 || requested_playlist > 99999 ||
        sample_bytes < 0 || sample_bytes > 6144) {
        io_error(env, "Invalid Blu-ray directory probe arguments");
        return NULL;
    }
    jclass cls = (*env)->GetObjectClass(env, provider);
    if (!cls) return NULL;
    SafTree ctx = {
        .env = env, .provider = provider,
        .open_fd = (*env)->GetMethodID(env, cls, "nativeOpenFd", "(Ljava/lang/String;)I"),
        .list_names = (*env)->GetMethodID(env, cls, "nativeListNames", "(Ljava/lang/String;)[Ljava/lang/String;"),
        .callback_error = 0,
    };
    (*env)->DeleteLocalRef(env, cls);
    if (!ctx.open_fd || !ctx.list_names) {
        io_error(env, "SAF Blu-ray JNI method resolution failed");
        return NULL;
    }

    BLURAY *bd = bd_init();
    if (!bd) { io_error(env, "Could not initialize libbluray"); return NULL; }
    jstring output = NULL;
    const char *failure = NULL;
    if (!bd_open_files(bd, &ctx, saf_open_dir, saf_open_file) || ctx.callback_error) {
        failure = "Could not open SAF Blu-ray BDMV folder through libbluray";
        goto cleanup;
    }
    const BLURAY_DISC_INFO *info = bd_get_disc_info(bd);
    if (!info || (info->aacs_detected && !info->aacs_handled) ||
        (info->bdplus_detected && !info->bdplus_handled)) {
        failure = "Protected Blu-ray folder cannot be read without authorized AACS/BD+ support";
        goto cleanup;
    }
    uint32_t count = bd_get_titles(bd, 0, 0);
    if (!count || count > 10000) {
        failure = "No native Blu-ray playlists found through SAF";
        goto cleanup;
    }
    uint32_t selected = UINT32_MAX;
    uint64_t duration = 0;
    for (uint32_t i = 0; i < count; ++i) {
        BLURAY_TITLE_INFO *title = bd_get_title_info(bd, i, 0);
        if (!title) continue;
        if (title->playlist <= 99999 && title->duration > 0 &&
            title->duration <= UINT64_C(90) * 1000 * 3600 * 48 &&
            ((requested_playlist == (int)title->playlist) ||
             (requested_playlist == -1 && title->duration > duration))) {
            selected = title->playlist;
            duration = title->duration;
        }
        bd_free_title_info(title);
    }
    if (selected == UINT32_MAX || !bd_select_playlist(bd, selected)) {
        failure = "Blu-ray playlist cannot be selected from SAF BDMV folder";
        goto cleanup;
    }
    int got = 0;
    if (sample_bytes) {
        uint8_t sample[6144];
        got = bd_read(bd, sample, sample_bytes);
        if (got <= 0) {
            failure = "libbluray cannot read M2TS payload from the SAF folder";
            goto cleanup;
        }
    }
    if (ctx.callback_error) {
        failure = "SAF provider failed during native Blu-ray reading";
        goto cleanup;
    }
    char report[160];
    int len = snprintf(report, sizeof(report),
        "MUKSMATT_ANDROID_BD_1\nP\t%u\t%llu\nR\t%d\n",
        selected, (unsigned long long)duration, got);
    if (len <= 0 || len >= (int)sizeof(report)) {
        failure = "Invalid native Blu-ray report";
        goto cleanup;
    }
    output = (*env)->NewStringUTF(env, report);

cleanup:
    bd_close(bd);
    if (failure) io_error(env, failure);
    return output;
}
