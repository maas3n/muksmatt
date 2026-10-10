# Phase 5 follow-up — Android Blu-ray to Matroska remux

Development PR. The merged PR #16 gave libbluray read-only ISO and BDMV file input through Android SAF; this PR adds a separate direct output engine.

## New processing path

1. Open the user-selected Blu-ray ISO SAF descriptor with `bd_open_stream()`, or the user-selected BDMV folder with existing `bd_open_files()` callbacks; select the requested playlist or longest viable playlist.
2. Read the native decrypted/clear playlist transport directly from libbluray through an FFmpeg custom `AVIOContext`. This is a one-pass path with no source.mkv staging. Do not use DVD-specific `100M / 100M / +genpts` flags.
3. Discover MPEG-TS streams and map all eligible video/audio/subtitles, or the explicitly selected input FFmpeg stream indices. Fail closed on invalid or duplicate selections.
4. Transfer native libbluray chapter start/end times (90 kHz) to Matroska chapter metadata, preserving chapters independent of the source MPEG-TS demuxer.
5. Every selected Blu-ray PCM_BLURAY stream **must** decode to lossless FLAC. Add only the PCM_BLURAY decoder, FLAC encoder and libswresample to the Android LGPL FFmpeg build. Copy all other compatible selected streams.
6. Write Matroska directly into a new SAF partial output. On a successful trailer, rename to the final `.mkv` filename; on failure, remove the partial document. No content is uploaded.

## Acceptance requirements / still to verify

- Native Android builds, Kotlin tests, ELF dependencies and Android emulator startup must pass for arm64-v8a and x86_64.
- Real-world BDMV and ISO should be verified for MPEG-TS probing, streams, M2TS timestamp continuity, chapter marks and lossless FLAC sample integrity. CI compilation alone does not prove remuxing works.
- AACS/BD+ remains fail-closed when not handled by installed authorized libraries; decryptor integration/keys/real media require separate validation.
- Physical USB Blu-ray drives, GUI tab hookups, playback parity, cancellation and progress may require further iterations; avoid advertising finished decryption or full end-user support on these foundations.

DVD JNI/stream flags, original MattRip, Android signing, releases and existing REMUX/DEMUX/ADVANCED/BATCH UI are unchanged.
