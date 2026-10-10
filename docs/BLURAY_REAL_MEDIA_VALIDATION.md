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
- PR #17 has already merged as a development foundation; keep physical-device and actual-media compatibility claims pending until dedicated evidence is collected
- AACS/BD+ and real USB drives are follow-on capabilities until validated; do not advertise them as tested or usable merely because libbluray is bundled

## Desktop CLI authored-media integration gate

The Linux workflow now executes `scripts/test-bluray-desktop-cli.sh` after building the actual Go CLI and native libbluray navigator. It authors an unencrypted 185-second BDMV folder and a UDF Blu-ray ISO using a checksum-pinned tsMuxer release (the separate Android C-core fixture remains six seconds), then exercises the **public** `muksmatt-cli bluray scan` and `bluray remux` entry points. Its coverage is distinct from the Android native C-core harness:

- scan folder and ISO independently, discover the longest playlist through Go and native libbluray respectively;
- remux both inputs to MKV, including video+audio, then perform explicit video-only and audio-only selections;
- verify MPEG-2 stream copy, mandatory 24-bit LPCM-to-FLAC output, chapter preservation and `--no-chapters`;
- hash all decoded source and output video frames and decoded 24-bit PCM samples, requiring exact parity;
- verify an invalid stream selection fails without publishing an MKV.

**Known FFmpeg limitation:** The desktop `bluray:` protocol enumerates only relevant playlists with a built-in 180-second minimum title length. This authored desktop fixture is deliberately 185 seconds; a playlist shorter than three minutes may still be discoverable with the muKsMaTT native navigator but fail in the FFmpeg remux path. Treat short-title remux as a separate open compatibility issue, not a validated feature. The native Android core does not use this `bluray:` protocol.

The test is *not* genuine commercial disc, drive, Windows execution, AACS/BD+, or Android/ChromeOS device evidence. Record its CI run result separately from the real-media qualifications below; do not mark this gate passed until its workflow completes successfully.

## Real-disc operator validation (still pending)

Use a legally accessible, non-sensitive test disc/backup and run with the installed **muKsMaTT** CLI and its bundled libbluray-enabled FFmpeg pair. Do not upload any copyrighted media or decryption material. For a local BDMV folder, ISO, or an accessible desktop optical drive, use:

```text
muksmatt-cli bluray scan SOURCE
muksmatt-cli bluray remux --output VALIDATION.mkv SOURCE
ffprobe -v error -show_entries stream=index,codec_type,codec_name,channels,sample_rate -show_chapters -of json VALIDATION.mkv
```

For Windows, replace the executable name with `muksmatt-cli.exe` and supply a valid Windows source path. Select another playlist explicitly with `--playlist 00800` only if that playlist appears in the scan. Do not run extraction with a non-writable path or to an existing file, and ensure ample free disk space. Visual playback and A/V sync still need a human/device check, especially at branching points and chapter transitions.

Save a private, sanitized report: OS/architecture; source class (folder/ISO/physical); reader model; playlist IDs/durations; selected stream indexes/codecs; input-versus-output stream mapping; chapter counts and timing; decoded hashes of owned test streams; playback and A/V sync observations; any AACS/BD+ protection *handled* or *unhandled* status. Remove disc titles, personally identifying paths, decrypted media, keys and passwords before sharing logs. Compare output to the original source rather than assuming that codec names alone establish lossless parity.

A genuine **physical-drive** validation remains outstanding on Windows and Linux; Android USB optical transport remains a separate, explicitly unverified milestone. The existing authored-fixture tests cannot clear those gates.
