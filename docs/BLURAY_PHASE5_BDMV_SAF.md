# Blu-ray Phase 5 — native Android BDMV folder SAF file bridge

Development-only follow-up to merged PR #15. DVD JNI/remux paths, CSS, signing, app UI and release workflow triggers remain unchanged.

## Native BDMV access

- Android `BluraySafTreeBridge` takes an explicitly user-granted SAF tree (disc root or BDMV directory), enumerates bounded child documents, indexes paths without exposing filesystem locations and rejects ambiguous/traversal paths.
- `BluraySafPathIndex` normalizes Blu-ray paths case-insensitively, rejects dot/dot-dot segments and unsafe names, and caps the number of documents. Valid BDMV/PLAYLIST and BDMV/STREAM directories must be present.
- Read-only `nativeOpenFd()` retrieves a descriptor only from the indexed user-selected SAF provider. The returned detached FD is owned by native `BD_FILE_H`, which closes it. Directory entries are served by an explicit `BD_DIR_H` callback; no intermediate .mkv or BDMV copy is produced.
- libbluray 1.5.0 `bd_open_files()` is used instead of a fake POSIX path. Read, seek, tell, close, eof and directory enumeration are supplied through SAF. A selected/longest playlist may be sampled for at most 6144 bytes, with AACS/BD+ handling checked.
- The existing Android seekable ISO path via `bd_open_stream()` is unchanged. The existing DVD libudfread 1.1.2 is untouched; native Blu-ray retains its private libudfread 1.2.0.
- Android/ChromeOS CI compiles both JNI entry points and runs pure Kotlin path-normalization tests; all existing Android native DVD regressions continue to run.

## Important limitations

This adds file-level Blu-ray BDMV reading and a bounded real transport probe, **not a Blu-ray-to-MKV muxer**. It does not yet offer Blu-ray stream selection, chapter mapping, LPCM-to-FLAC, remux progress, a DEMUX tab, or an Android Blu-ray GUI action.

Android libaacs/libbdplus decryptors, authorized disc key/configuration sources, optical USB drive access, genuine media fixture tests and device validation are not part of this change. Protected discs whose encryption is not handled continue to fail closed. CI build success is not an encrypted Blu-ray remux verification.

Reference: https://videolan.videolan.me/libbluray/bluray_8h.html (bd_open_files).
