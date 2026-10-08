# Blu-ray Phase 5 — Android/ChromeOS native ISO reader foundation

**Development milestone, not release-ready.** DVD JNI and REMUX/DEMUX/ADVANCED/BATCH UI remain unchanged.

## Implemented

- Seekable Android SAF ISO 2048-byte block reader using pread(), with strict bounds, descriptor duplication and pipe/truncated-image rejection.
- Separate libmuksmatt_bluray.so JNI reader linked with libbluray 1.5.0; invokes bd_open_stream(), scans native playlists, and optionally reads up to 6144 video transport bytes.
- Kotlin BlurayNativeIsoBridge takes a read-only SAF file descriptor, keeps it open during the native operation and strictly validates native scan results.
- arm64-v8a and x86_64 builds with pinned upstream libbluray 1.5.0 source checksum f676408e91a5d321abf8b8d4dfdae36205c297dab5c54c3ec519639025f474a2, and 16 KB ELF segment alignment. The private Blu-ray runtime uses independently packaged libudfread 1.2.0 with a distinct SONAME; the existing DVD libudfread 1.1.2 is unchanged.
- Source archive and license notices, Android APK/AAB audit, host block-reader regression tests and Kotlin native metadata parsing tests.

## Safety

- Does not modify DVD JNI or use unrestricted Android filesystem paths.
- Native code duplicates and owns the borrowed SAF descriptor and closes it when libbluray finishes.
- Protected sources fail when AACS or BD+ cannot be handled by the native environment. No keys, certificates or decryption data are downloaded or bundled.

## Not yet implemented

1. Native BDMV folder traversal with SAF callbacks via libbluray bd_open_files(); the earlier Android SAF implementation currently reads MPLS metadata only.
2. End-to-end Blu-ray MKV remux, stream inventory, chapter injection and Blu-ray LPCM -> FLAC conversion. This phase adds a bounded real transport read but not a muxer.
3. Built-in Android libaacs/libbdplus and verified encrypted-disc decryption.
4. Portable USB optical drive transport, Android Blu-ray UI actions, and physical-device validation.
5. Actual ISO/BDMV and encrypted-disc test media and playback/remux quality verification.

PR stays draft/unmerged until CI is verified, and public releases remain disabled. MattRip upstream is unchanged.

References: https://videolan.videolan.me/libbluray/bluray_8h.html and https://get.videolan.org/libbluray/last/libbluray-1.5.0.tar.xz
