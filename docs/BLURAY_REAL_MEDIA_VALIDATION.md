# Phase 5 — physical Blu-ray and Android SAF validation matrix

## Validated with reproducible authored media

- Host-native integration: BDMV directory and UDF ISO; direct MKV output, chapter presence, MPEG-2/H.264 decoded-frame hashes, AC-3 elementary stream byte parity, 24-bit stereo and six-channel LPCM-to-FLAC exact decoded sample parity
- Device-path CI: debug-only authored BDMV and ISO fixtures, actual Android DocumentsProvider file/directory callbacks, native libbluray JNI, SAF output document creation and final rename, selective video/audio remux, chapters, and rollback of incomplete outputs on API 26 / API 35 / API 35 16 KB emulator images
- Status is **pending** until the new CI emulator test jobs actually pass; not equivalent to a physical disc test

## Still necessary before claiming real-disc compatibility

| Source/media | Test | Required evidence |
|---|---|---|
| Unencrypted pressed Blu-ray or authorized full-disc backup | Physical optical drive and equivalent BDMV + ISO | Correct longest-title discovery, chapters and final playback |
| AACS-protected disc with a legally authorized available key/configuration | libaacs and libbluray decryption on Windows, Linux and Android where available | Native disc info reports protection *handled*, no failed reads, decrypt result matches selected playlist |
| BD+ protected title | libbdplus integration on supported hosts | BD+ handled, picture/audio correct, no unreported protection bypass |
| Multi-playlist / seamless branching | Select individual playlists across clip boundaries | Title/playlist maps, chapter start/end, audio and video sync across transitions |
| HEVC/H.264/MPEG-2, VC-1 | Stream-copy and playback | Exact decoded video, no transcoding, correct HDR metadata where applicable |
| TrueHD/Atmos, DTS-HD MA/DTS:X, AC-3/E-AC-3 | Audio stream copy | Exact encoded core/extension streams and language mapping; no silent downmix |
| LPCM 16/20/24-bit, stereo/5.1/7.1 | Mandatory lossless FLAC on each platform | Sample-by-sample parity and preserved channel order, sample rates and bit depth |
| PGS subtitle and multiple tracks | Stream selection, subtitles in MKV | Correct palette, timing, forced/default flags and playback |
| USB BD reader on Android/ChromeOS | USB OTG with power and permission | SAF browse/input or documented alternative, sustained full-disc reading |
| Multiple SAF providers (local files, DocumentsUI, removable USB) | Non-path-based access and output, cancellation | Correct descriptor lifetime, no staging source.mkv, no partial MKV after failure |

## Collection and safety

These tests cannot be fully performed on GitHub-hosted emulators: they do not have a physical BD optical drive, protected commercial media, device-owned external USB storage or necessary personal decryption keys. Do not add copyrighted full discs, decryption secrets, or signing materials to public GitHub artifacts. Obtain authorization for the source media and comply with local law.

For real-disc qualification, collect *non-sensitive* evidence: source type, playlist ID, stream codecs/counts, duration, chapter count, selected output streams, FFmpeg/ffprobe result summary, Blu-ray protection handled/unsupported flag, hash comparisons of user-authorized sample streams, device/OS/architecture, and provider type. A 100% green CI build alone is not sufficient.

## Merge/release gates

- A successful emulator SAF remux of both self-authored BDMV and ISO is a gate for the Android native remux implementation
- Verify no DVD remux/demux regression, no changes to DVD-specific FFmpeg input policy, and no release-signing or publishing regression
- Keep PR #17 draft until documented high-priority device and actual-media coverage is sufficient for its intended scope
- AACS/BD+ and real USB drives are follow-on capabilities until validated; do not advertise them as tested or usable merely because libbluray is bundled
