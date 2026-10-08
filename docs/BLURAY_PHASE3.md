# Blu-ray Phase 3 — native ISO/drive and Android SAF metadata

Development milestone only. Existing DVD engines, Android JNI and UI layout remain unchanged. Public releases stay disabled.

## Desktop
The Go BDMV folder parser remains intact. For ISO images and optical drives, a separately compiled native libbluray companion named muksmatt-bluray-nav enumerates actual titles and chapter metadata using bd_open, bd_get_titles and bd_get_title_info. It uses native Blu-ray 90 kHz timing converted into the common 45 kHz chapter model.

Physical-drive detection rejects ordinary non-optical block devices. Linux uses /dev/sr*; Windows verifies the Win32 optical drive type. FFmpeg's bluray protocol reads the selected playlist directly into MKV. Blu-ray LPCM is converted to lossless FLAC and other compatible tracks are stream-copied. DVD-only probe flags are never inserted on these inputs.

Linux CI now compiles the helper against libbluray-dev and checks its protocol version. Linux packaging copies the companion and declares libbluray2 as a dependency. Windows Go detection and runtime adapters compile, but **a Windows native helper and DLL runtime are not yet bundled**; do not claim Windows ISO/optical support is enabled in the stock application. The bundled FFmpeg and FFprobe must also expose the bluray protocol, otherwise runtime fails explicitly.

## Android / ChromeOS
BluraySafReader implements real read-only DocumentsContract traversal for an explicitly granted root/BDMV tree. It enumerates the PLAYLIST directory and reads numbered MPLS files with a 16 MiB per-file bound, then passes them to the Kotlin playlist catalog and 45 kHz chapter parser. The DVD JNI engine is never invoked.

This is navigation metadata reading only. A seekable native libbluray and native libaacs/libbdplus reader, Android Blu-ray UI, encrypted M2TS reading, ISO reading and physical USB optical-drive access are NOT implemented yet.

## Encryption and testing
libbluray can use locally available libaacs and libbdplus, with user-authorized configuration. The companion refuses protected sources whose protection cannot be handled. No decryption keys, certificates or BD+ VM data are fetched, bundled or uploaded. A CI-green build is not proof that encrypted Blu-rays are decryptable.

Still required: bundled native helper and native libraries on Windows and Android; FFmpeg libbluray packaging; synthetic unencrypted ISO integration tests; real AACS/BD+ discs; optical-drive/device tests; full chapters/audio/subtitle parity; and physical Android/ChromeOS tests. No real protected physical Blu-ray disc was available for validation in this phase.

## CLI examples
    muksmatt-cli bluray scan /path/to/BDMV-root
    muksmatt-cli bluray scan /path/to/movie.iso
    muksmatt-cli bluray scan /dev/sr0
    muksmatt-cli bluray remux --playlist 00800 --output movie.mkv /path/to/movie.iso
    muksmatt-cli bluray remux --output movie.mkv /dev/sr0

References:
- libbluray API: https://videolan.videolan.me/libbluray/bluray_8h.html
- FFmpeg bluray protocol: https://ffmpeg.org/doxygen/9.0/bluray_8c.html
