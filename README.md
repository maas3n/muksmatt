# muKsMaTT

**muKsMaTT** is a cross-platform media remuxing, demuxing, merging, batch-processing, and CLI application for **Windows, Linux, Android, and ChromeOS**. It is developed from [MattRip beta-build-6-1](https://github.com/maas3n/MattRip/releases/tag/beta-build-6-1), preserving the existing interface, DVD processing, and direct stream demux.

> **Public beta — v0.1.0-beta.1 (experimental).** The first muKsMaTT GitHub prerelease is available for Windows, Linux, Android, and ChromeOS community testing. It is not a stable production release. Physical Blu-ray drives, protected Blu-ray discs, and real-device compatibility still require testing. Use the muKsMaTT release downloads below rather than MattRip packages or CI development artifacts.

The long-term objective is **DVD and Blu-ray remux/demux and decryption** on Windows, Linux, and Android, without transcoding existing video/audio/subtitle streams where stream-copy is feasible. **Experimental Blu-ray development is implemented for authored BDMV folders and ISO media, but has not been validated on real physical Blu-ray drives or protected commercial discs.** Android physical USB optical-drive handling is a separate, unverified milestone. The current DVD build uses existing libdvdread/libdvdnav/libdvdcss integrations, with the Google Play variant deliberately CSS-free pending a distribution review.

## Download v0.1.0-beta.1 (public beta)

The first public muKsMaTT release is an **experimental GitHub prerelease**. Click a package below to download its file directly (not just open the release page):

| Platform | Direct download |
| --- | --- |
| **Android / ChromeOS** (Android 8.0+, where APK sideloading is supported) | [Download Android APK](https://github.com/maas3n/muksmatt/releases/download/v0.1.0-beta.1/muKsMaTT-0.1.0-beta.1-Android.apk) |
| **Linux** (x86-64, glibc 2.38+) | [Download Linux Standalone](https://github.com/maas3n/muksmatt/releases/download/v0.1.0-beta.1/muKsMaTT-0.1.0-beta.1-Linux-amd64Standalone) |
| **Windows** (x64) | [Download Windows All-in-One](https://github.com/maas3n/muksmatt/releases/download/v0.1.0-beta.1/muKsMaTT-0.1.0-beta.1-Windows-All-in-One.exe) |

[Full release, alternative installers and sources](https://github.com/maas3n/muksmatt/releases/tag/v0.1.0-beta.1) · [SHA-256 checksum file](https://github.com/maas3n/muksmatt/releases/download/v0.1.0-beta.1/SHA256SUMS.txt) · [Beta testing guide](docs/releases/v0.1.0-beta.1-testing.md)

### Direct downloads with `wget` or `curl`

Both tools follow GitHub's release-download redirects. Run these examples in a Linux shell (including Debian WSL). You only need **one** download method, and may choose only the package you need.

**Using `wget`:**

```bash
wget -O "muKsMaTT-0.1.0-beta.1-Android.apk" "https://github.com/maas3n/muksmatt/releases/download/v0.1.0-beta.1/muKsMaTT-0.1.0-beta.1-Android.apk"
wget -O "muKsMaTT-0.1.0-beta.1-Linux-amd64Standalone" "https://github.com/maas3n/muksmatt/releases/download/v0.1.0-beta.1/muKsMaTT-0.1.0-beta.1-Linux-amd64Standalone"
wget -O "muKsMaTT-0.1.0-beta.1-Windows-All-in-One.exe" "https://github.com/maas3n/muksmatt/releases/download/v0.1.0-beta.1/muKsMaTT-0.1.0-beta.1-Windows-All-in-One.exe"
```

**Using `curl`:**

```bash
curl -fL -o "muKsMaTT-0.1.0-beta.1-Android.apk" "https://github.com/maas3n/muksmatt/releases/download/v0.1.0-beta.1/muKsMaTT-0.1.0-beta.1-Android.apk"
curl -fL -o "muKsMaTT-0.1.0-beta.1-Linux-amd64Standalone" "https://github.com/maas3n/muksmatt/releases/download/v0.1.0-beta.1/muKsMaTT-0.1.0-beta.1-Linux-amd64Standalone"
curl -fL -o "muKsMaTT-0.1.0-beta.1-Windows-All-in-One.exe" "https://github.com/maas3n/muksmatt/releases/download/v0.1.0-beta.1/muKsMaTT-0.1.0-beta.1-Windows-All-in-One.exe"
```

**Verify downloaded files** against the release's SHA-256 checksums (Linux / WSL):

```bash
curl -fL -o SHA256SUMS.txt "https://github.com/maas3n/muksmatt/releases/download/v0.1.0-beta.1/SHA256SUMS.txt"
sha256sum --ignore-missing -c SHA256SUMS.txt
```

To launch the Linux Standalone, make it executable first: `chmod +x muKsMaTT-0.1.0-beta.1-Linux-amd64Standalone`. The links intentionally target the **published beta tag** rather than `releases/latest/download/` (which does not point to GitHub prereleases). When a newer beta is published, its versioned filenames and links will need updating.

## Build and release status

- [GitHub Actions and development artifacts](https://github.com/maas3n/muksmatt/actions)
- [Open development pull requests](https://github.com/maas3n/muksmatt/pulls)
- [Published v0.1.0-beta.1 prerelease](https://github.com/maas3n/muksmatt/releases/tag/v0.1.0-beta.1) — signed Android APK, Linux and Windows packages, sources, notices, and SHA-256 checksums.
- [Public beta release notes](docs/releases/v0.1.0-beta.1.md) and [community testing guide](docs/releases/v0.1.0-beta.1-testing.md) — known limitations and reporting instructions.

The public beta uses **versioned asset filenames** (for example, `muKsMaTT-0.1.0-beta.1-Android.apk`). Use the versioned links in the download section above for both browser and command-line downloads; `releases/latest/download/` does not target this GitHub prerelease. The portable Windows executable and CLI use lowercase `muksmatt.exe` and `muksmatt-cli.exe`.

## What muKsMaTT can do

muKsMaTT keeps the established DVD and general-media workflows while adding **experimental Blu-ray support**. Supported video, audio, and subtitle streams are remuxed into Matroska (**MKV**) without a lossy re-encode. An important exception is **Blu-ray LPCM**: selected `pcm_bluray` audio is deliberately encoded as **lossless FLAC**, retaining decoded audio samples rather than passing the raw LPCM format through unchanged. The exact behavior depends on the source and workflow.

### Blu-ray — experimental DVD-independent path

Blu-ray support on Windows and Linux is available in **REMUX/DEMUX**, **ADVANCED**, and the in-app **CLI** tab (as well as the packaged desktop CLI). Android/ChromeOS has a separate experimental native SAF remux implementation. The desktop Blu-ray reader follows a selected MPLS playlist using libbluray and supports direct stream extraction without a temporary MKV.

The desktop CLI recognizes accessible **BDMV disc-root folders**, **Blu-ray ISO images** and **optical-drive paths** where the OS grants access. It can scan MPLS playlists, display track PIDs/languages/codecs, directly DEMUX selected elementary streams (or video-only M2TS), and REMUX to MKV or M2TS. Use **--playlist** to explicitly select the correct title: deliberately looping or obfuscated playlists can be longer than the movie. MKV can preserve chapters and losslessly convert Blu-ray LPCM to FLAC; M2TS stream-copies compatible video/audio/PGS into one container and excludes chapters. Blu-ray CLI uses the same FFmpeg/FFprobe/libbluray tools as the GUI.

**Desktop CLI examples** (use `muksmatt-cli.exe` on Windows, included in the Setup and Portable packages):

```bash
# List MPLS playlists, then inspect the correct one
muksmatt-cli bluray scan "/path/to/bluray-disc-root"
muksmatt-cli bluray tracks --playlist 00800 "/path/to/bluray-disc-root"

# Remux selected Blu-ray streams into MKV, retaining playlist chapters
muksmatt-cli bluray remux --playlist 00800 --output movie.mkv "/path/to/bluray-disc-root"

# M2TS combines compatible video, audio and PGS (no chapters)
muksmatt-cli bluray remux --playlist 00800 --output movie.m2ts "/path/to/bluray-disc-root"

# Directly extract H.264 and a selected DTS/PGS stream; no temporary MKV
muksmatt-cli bluray demux --playlist 00800 --output "/existing/output/folder" --streams 0,1,9 "/path/to/bluray-disc-root"

# Alternatively demux both native video (.h264/.hevc) and video-only .m2ts
muksmatt-cli bluray demux --playlist 00800 --output "/existing/output/folder" --video both "/path/to/bluray-disc-root"
```

Substitute an accessible `.iso` or supported optical-drive path for the disc root. Existing output files are not overwritten; for DEMUX, `--output` names an **existing directory** and the application creates its own extraction subfolder. The in-app **CLI** tab offers Blu-ray Scan / Tracks / REMUX / DEMUX templates; replace `SOURCE` and `OUTPUT_DIR` before running. On Linux the Standalone also accepts the `--cli` option.

**Blu-ray limitations:** CI has exercised *self-authored, unencrypted* BDMV and ISO fixtures, chapter mapping, stream selection, and LPCM-to-FLAC parity; genuine physical Blu-ray drives, protected commercial discs, real-device Android compatibility and **AACS/BD+ decryption** remain unverified. Bundled `libaacs`/`libbdplus` libraries do **not** imply working protection handling or bundled keys. The desktop FFmpeg `bluray:` protocol can reject playlists shorter than **180 seconds** ([issue #20](https://github.com/maas3n/muksmatt/issues/20)). Android USB optical-drive input is not a confirmed supported feature.

For the actual qualification status, see the [Blu-ray real-media validation matrix](docs/BLURAY_REAL_MEDIA_VALIDATION.md), [beta release notes](docs/releases/v0.1.0-beta.1.md), and [community testing guide](docs/releases/v0.1.0-beta.1-testing.md).

### DVD and MKV — REMUX/DEMUX

The main **REMUX/DEMUX** tab handles DVD, MKV and experimental Blu-ray sources. It accepts:

- DVD folders and `VIDEO_TS`
- unmounted DVD ISO images
- physical DVD drives on Windows/Linux (`D:`-style drive sources on Windows and `/dev/sr0`-style devices on Linux)
- MKV files

For DVD sources, muKsMaTT discovers DVD titles, automatically selects the longest readable title, and lets you inspect metadata before processing. Video, audio, and subtitle streams are individually selectable. All detected streams are selected by default.

Choose **REMUX** to create an MKV using stream copy. Chapter preservation is optional.

Windows and Linux use FFmpeg's `dvdvideo` input backed by `libdvdread`/`libdvdnav`. Self-contained desktop builds keep libdvdcss private to muKsMaTT and let libdvdread discover it dynamically; no separate CSS remux engine is introduced. Direct GitHub/development Android builds statically link libdvdcss into the JNI bridge and decrypt scrambled VOB sectors before the existing native FFmpeg/libav stream-copy reader consumes them. The Google Play build keeps CSS disabled for now. muKsMaTT does not use its own DVD IFO parser.

### DVD/MKV demux

The **Demux** action in the **REMUX/DEMUX** tab extracts selected streams without re-encoding. It works with DVD folders, DVD ISOs, and MKV sources on all supported platforms, plus physical DVD drives on Windows/Linux.

Depending on the selected streams, muKsMaTT can export:

- MPEG-2 video as `.mpeg2` or video-only `.VOB`
- H.264 as `.h264`
- HEVC/H.265 as `.h265`
- audio in its codec format, including AC-3 and DTS
- SRT and ASS text subtitles
- PGS subtitles as `.sup`
- DVD/VobSub subtitles as paired `.idx` + `.sub`
- chapters as `Chapters.txt` in simple OGM chapter format

The VOB option creates a video-only MPEG program stream; it does not recreate DVD menus or DVD-Video structure. Unsupported selected codecs are reported rather than transcoded.

Each demux operation creates a new output folder. DVD folders and ISOs are read directly from the selected title rather than being remuxed to a temporary MKV first. DVD inputs use the `100M / 100M / +genpts` input policy (`-analyzeduration 100M -probesize 100M -fflags +genpts` on desktop, with the native libav equivalent on Android/ChromeOS). MKV demux keeps its container-timestamp path instead of inheriting the DVD-specific 100M overrides.

Android/ChromeOS temporarily holds exported files while saving them through the Storage Access Framework; MKV inputs also require an app-private input copy. Direct DVD demux keeps stream selection, chapters, subtitle metadata, cancellation, extraction progress, and the DVD clock-reset/discontinuity handling used by the native reader.

### ADVANCED (general media and DVD)

The **ADVANCED** tab combines selected streams from multiple sources into one MKV without transcoding.

Inputs include normal containers such as MKV, MP4 and AVI, DVD ISOs, **physical DVD drives on Windows/Linux**, and supported elementary streams such as H.264, MPEG-2/VOB, AAC, AC-3, MP3, DTS, SRT, WebVTT and SUP.

- **MOVIE FILES** exposes discovered video, audio, subtitle, attachment/data streams and embedded chapters.
- **DVD DRIVE** adds the longest readable title from a physical desktop DVD drive.
- **AUDIO / MKV / RAW** exposes audio streams.
- **SUBTITLE / MKV / RAW** exposes subtitle streams.
- **CHOOSE CHAPTER FILE FROM MKV or RAW** accepts chapters from MKV or valid `FFMETADATA1`.

Streams are explicitly selectable. Embedded chapter titles are preserved, and a dedicated chapter source overrides selected embedded chapters.

The **ADVANCED** tab also has a **DEMUX** action. Tick or untick the rows you want, then demux only the selected video, audio, and subtitle streams; a checked embedded chapter row exports `Chapters.txt`. Multiple inputs can be demuxed in one operation, with source-prefixed output names to avoid collisions. MPEG-2 video can be exported either as elementary `.mpeg2` or as video-only `.VOB`, matching the main Demux workflow.

DVD ISO and physical-drive input use the longest readable DVD title and keep that title through stream selection and muxing. Windows/Linux Advanced Merger DEMUX reads DVD selections directly through `dvdvideo`. Android/ChromeOS maps the selected staged merger rows back to the original DVD stream indexes and performs DEMUX directly from the original DVD source rather than from the temporary MKV used for the MUX workflow. Android/ChromeOS stages Storage Access Framework documents as needed; paired VobSub input requires both the matching `.idx` and `.sub` files.

See [`docs/ADVANCED_MERGER.md`](docs/ADVANCED_MERGER.md) for details.

### BATCH

**BATCH REMUX** scans a collection and losslessly remuxes discovered DVD movies to MKV. It supports `Movie/VIDEO_TS` folders and unmounted `.iso` files, automatically selects the longest readable title, preserves streams and chapters, and continues to later items when one item fails.

Example collection:

```text
Movies/
├── Alien.iso
├── Movie One/
│   └── VIDEO_TS/
│       ├── VIDEO_TS.IFO
│       └── ...
└── Movie Two/
    └── VIDEO_TS/
        ├── VIDEO_TS.IFO
        └── ...
```

Without an explicit output root, an ISO such as `Alien.iso` produces `Alien.mkv` beside the ISO, while `Movie One/VIDEO_TS` produces `Movie One/Movie One.mkv`. An explicit output root collects completed MKVs in that destination. Existing MKVs are not overwritten.

Windows and Linux also expose batch processing through the packaged CLI. Android/ChromeOS provides BATCH through its Storage Access Framework-based app implementation.

### CLI

Windows and Linux share the core CLI commands:

```text
scan
metadata
remux
--batch
```

Single-disc remuxing supports options including `--title`, `--streams 0,2`, `--no-chapters`, and `--output`. The source can be a DVD folder, ISO image, or physical DVD device such as `D:` on Windows or `/dev/sr0` on Linux. Physical drives require an explicit `--output` because there is no writable folder beside the disc. Windows Setup/Portable packages include `muksmatt-cli.exe`; Linux packages include `muksmatt-cli`, and the Linux Standalone accepts CLI commands directly or after `--cli`.

Example:

```bash
muksmatt-cli remux --title 1 --streams 0,2 /path/to/DVD-or.iso
muksmatt-cli remux --output /path/to/output D:
muksmatt-cli remux --output /path/to/output /dev/sr0
muksmatt-cli --batch --log=/path/to/muksmatt-batch.log /path/to/Movies /path/to/output
```

Android/ChromeOS also includes an in-app CLI with `scan`, `metadata`, `remux`, and `--batch` using Storage Access Framework content URIs. Android remux supports explicit `--streams` selection and `--no-chapters`.

**BATCH remains DVD-oriented.** On Windows and Linux, the CLI supports DVD and Blu-ray operations; Blu-ray video, audio and PGS tracks can be extracted directly without constructing an intermediate MKV. For mixed-source MUX with editable imported-track metadata, use ADVANCED.

## Track selection and metadata

Use **SCAN/SELECT STREAMS** to scan a new source (automatically selecting the longest DVD title) and choose exactly which video, audio, and subtitle streams to process. After changing the title selection, use the same button to refresh the stream selection for that title.

For example:

```text
Video
☑ #0  MPEG-2 Video   720×576

Audio
☑ #1  AC-3 5.1      English
☐ #2  AC-3 2.0      Commentary

Subtitles
☑ #3  DVD Subtitle  English
☐ #4  DVD Subtitle  Norwegian
```

All tracks start selected. At least one media stream must remain selected for remuxing or demuxing. Chapter preservation/extraction is controlled separately.

Changing the source or DVD title clears the previous selection so stream indexes cannot accidentally carry over to another title. MKV metadata in the **REMUX/DEMUX** tab is displayed using MediaInfo.

## Platform support and compatibility

| Feature | Windows | Linux | Android / ChromeOS |
| --- | --- | --- | --- |
| DVD `VIDEO_TS` folders and DVD ISO remux/demux | Yes | Yes | Yes, via SAF |
| Physical DVD drive input | Supported desktop path | Supported desktop path | USB optical drive **not validated** |
| MKV demux / general-media Advanced Merger | Yes | Yes | Yes |
| Blu-ray BDMV / ISO remux | **Experimental desktop CLI** | **Experimental desktop CLI** | **Experimental native SAF path**; GUI integration not established |
| Blu-ray playlist selection / chapters / LPCM→FLAC | Experimental CLI | Experimental CLI | Experimental native path; verify on real devices |
| Physical Blu-ray drives | Implementation exists, **hardware unverified** | Implementation exists, **hardware unverified** | **Unverified** |
| AACS / BD+ protected Blu-ray decryption | **Not validated** | **Not validated** | **Not validated** |

A green build or emulator test does not establish compatibility with real, encrypted, or commercially authored Blu-rays. Please check the [beta testing guide](docs/releases/v0.1.0-beta.1-testing.md) before reporting success or failure.

### Windows

- Windows x64
- All-in-One single-file GUI
- normal Setup package
- Portable ZIP
- packaged `muksmatt-cli.exe` in Setup/Portable
- bundled FFmpeg, FFprobe and MediaInfo
- current binaries are not Authenticode-signed

### Linux

- amd64 / x86_64
- one-file Standalone executable
- self-contained `.deb` and tarball packages
- GUI and packaged CLI
- published binaries currently require **glibc 2.38 or newer**
- Standalone includes private graphics/runtime dependencies while still relying on the host glibc, display session and GPU driver environment

### Android / ChromeOS

- Android 8.0 / API 26 or newer
- arm64-v8a and x86_64 native runtimes
- one persistently signed universal APK for Android phones/tablets and Chromebooks with Android app support
- native stream-copy DVD remux, direct DVD/MKV demux, Advanced Merger MUX/DEMUX, BATCH and in-app CLI paths
- Storage Access Framework input/output
- DVD folder and read-only UDF ISO support
- CI coverage includes API 26, API 35 and API 35 with 16 KB pages

Android/ChromeOS remains **experimental** while physical-device, Chromebook, Blu-ray hardware and Play Store validation remain separate gates. The GitHub beta APK is signed with muKsMaTT's **own persistent app-signing identity** and uses application ID `io.github.maas3n.muksmatt`; it is a separate app from MattRip and MattMux, not an in-place upgrade for their APKs. Google Play purchases and public Play distribution are not enabled.

See [`android/README.md`](android/README.md) for Android-specific implementation details and limitations.

## Important limitations

- CSS support does not add Android USB optical-drive transport; Android CSS-capable builds currently apply to supported DVD folder/ISO inputs through the existing SAF/UDF path.
- Interleaved multi-angle DVD titles are unsupported on the current Android path.
- Still/shuffle/multi-PGC DVD semantics are not fully supported on Android.
- Android ISO input requires a seekable storage provider.
- Android operations that stage media require sufficient temporary free space.
- Output capabilities can depend on the selected Android document provider.
- Remuxing and demuxing are stream-copy operations; unsupported formats are not silently transcoded.

## Install and get started

Use the [verified direct beta downloads](#download-v010-beta1-public-beta) above. These are public **prerelease** packages; always keep the original source media and verify the [release SHA-256 checksums](https://github.com/maas3n/muksmatt/releases/download/v0.1.0-beta.1/SHA256SUMS.txt).

- **Windows x64:** Download and run the All-in-One `.exe`, or choose Setup / Portable from the release page. The Windows executables are **not Authenticode-signed**, so Windows may display a publisher warning. Check the checksum and source before choosing to run an unsigned beta. The desktop CLI is included with Setup/Portable.
- **Linux x86-64:** The Standalone requires **glibc 2.38+** and a usable graphical session to launch the GUI. For example:

  ```bash
  chmod +x muKsMaTT-0.1.0-beta.1-Linux-amd64Standalone
  ./muKsMaTT-0.1.0-beta.1-Linux-amd64Standalone
  ```

  The release also provides a `.deb` and Linux tarball.
- **Android / ChromeOS:** Install the signed APK using your device's supported sideloading method. Android 8.0/API 26+ is required; Chromebooks must support Android applications. This is not a Google Play download. Grant access to your chosen files/folders through Android's document picker (SAF); seekable storage providers and sufficient temporary space may be needed.

**First DVD/MKV remux:** In **REMUX/DEMUX**, select a DVD `VIDEO_TS` folder, DVD ISO or MKV, choose **SCAN/SELECT STREAMS**, review the title/tracks and chapter setting, select a writable output destination, then choose **REMUX**. Use **DEMUX** to extract selected streams. **ADVANCED** combines selected streams from multiple sources, and **BATCH REMUX** processes a DVD collection.

**First Blu-ray test:** Use the [desktop CLI examples above](#blu-ray--experimental-dvd-independent-path) with a legally accessible, preferably unencrypted BDMV/ISO. Compare resulting MKV streams, chapters, and audio/video sync with the source. Android Blu-ray support remains experimental and device/provider dependent; avoid assuming that the DVD GUI controls perform Blu-ray remux.

Consult the [community testing guide](docs/releases/v0.1.0-beta.1-testing.md) before testing. The build-from-source instructions below track `main`.

## Troubleshooting and beta feedback

- **Blu-ray scan or remux fails:** Check whether the selected source is an accessible BDMV disc root, an actual Blu-ray ISO, or a readable optical device. Protected discs may require separately configured, legally available information; **AACS/BD+ has not been validated**. Playlists shorter than three minutes are a known desktop limitation.
- **Linux application won't launch:** Check `chmod +x`, the host `glibc` version (`ldd --version`), and whether a graphical display session is available.
- **Android source/output inaccessible:** Re-select it using the system file/folder picker. Some document providers cannot seek or rename output documents; temporary staging also requires free device space.
- **Unexpected streams, chapters, or lip-sync:** Compare the resulting MKV with the source; note the selected playlist/title, audio/subtitle stream indexes, chapter count, and where playback diverges.

Please [report bugs or successful compatibility tests through GitHub Issues](https://github.com/maas3n/muksmatt/issues/new). Include the beta version, OS/architecture, package, source type (DVD/BDMV/ISO/optical drive), selected streams, expected vs actual result, steps to reproduce and sanitized log messages. The [copy-ready community test report](docs/releases/v0.1.0-beta.1-testing.md) lists further details. **Do not upload copyrighted disc content, decryption keys, secrets, private paths or personal identifiers.**

## Build from source

Development builds use `main`; published binaries use the immutable version tag.

### Windows

```powershell
powershell -ExecutionPolicy Bypass -File .\src\build.ps1
```

### Linux

```bash
bash packaging/linux/build-linux-release.sh dev
bash packaging/linux/build-linux-standalone.sh dev
```

### Android / ChromeOS

Build the native runtime first, then run the Android tests/lint/package build:

```bash
export ANDROID_NDK_HOME="$ANDROID_HOME/ndk/30.0.16248370"
bash android/native/build-ffmpeg-android.sh
gradle -p android :app:testDebugUnitTest :app:lintDebug :app:assembleDebug
```

GitHub Actions uses the same native build step before Gradle. See [`android/native/README.md`](android/native/README.md) for the full native build process.

## Development and release model

`main` is the only long-lived source branch. Windows, Linux, and Android/ChromeOS changes are integrated into the same trunk and tested together.

New public versions use one shared tag:

- stable: `vMAJOR.MINOR.PATCH`
- preview: `vMAJOR.MINOR.PATCH-alpha.N`, `-beta.N`, or `-rc.N`

A release is built from one source commit and publishes the applicable Windows, Linux and Android/ChromeOS packages together. Published tags and release assets are treated as immutable; fixes are shipped as a new version.

See [`RELEASING.md`](RELEASING.md) for the full release policy.

## Recent development

The current feature set grew substantially after the early 1.4.x releases. Notable additions and fixes include cross-platform Advanced Merger expansion, Advanced Merger selected-stream DEMUX with MPEG-2/VOB choice, one-click BATCH, shared Windows/Linux DVD CLI commands, Android BATCH and in-app CLI support, DVD ISO handling, native Android libdvdnav title selection, MKV input in the REMUX/DEMUX tab, direct DVD demux without a temporary MKV, DVD clock-reset/progress handling, DVD subtitle extraction, and stronger Windows/Linux/Android parity coverage.

muKsMaTT was forked from [MattRip beta-build-6-1](https://github.com/maas3n/MattRip/releases/tag/beta-build-6-1), which originated from MattMux. **muKsMaTT has its own versioned releases, build outputs and Android application/signing identity.** Earlier MattRip/MattMux releases are historical provenance, not muKsMaTT downloadable packages; see [MattMux Releases](https://github.com/maas3n/MattMux/releases) for earlier history.

## Third-party runtime tools

Desktop builds use FFmpeg/FFprobe and MediaInfo CLI. Android/ChromeOS uses native FFmpeg/libav plus DVD/UDF libraries and MediaInfo where required by the feature path.

Exact pinned versions, hashes, source revisions, licensing notes, and provenance are documented in [`THIRD_PARTY.md`](THIRD_PARTY.md).

## License

muKsMaTT is licensed under the [MIT License](LICENSE). Third-party components remain governed by their own licenses.

---

## Complete CLI guide (Windows and Linux)

This step-by-step guide covers the **CLI tab inside muKsMaTT** and the standalone command-line executable on **Windows and Linux**. Begin with the basics, then use the advanced examples for precise DVD titles, Blu-ray playlists, selected tracks, lossless MKV/M2TS output, and direct Blu-ray DEMUX. This section documents the desktop CLI in the current development branch; some commands may not exist in older public beta packages. Android/ChromeOS uses a separate SAF-based CLI implementation.

### 1. Getting started: where to enter commands

There are two supported ways to use the desktop CLI:

1. **Inside muKsMaTT:** Open the **CLI** tab, type one command in its command field, and click **Run**. The Blu-ray **Scan**, **Tracks**, **REMUX**, and **DEMUX** buttons insert editable example commands. Replace `SOURCE`, `OUTPUT_DIR`, and other placeholders with real paths before running. **Cancel** requests cancellation of the running task.
2. **In a terminal:** Use `muksmatt-cli` on Linux or `muksmatt-cli.exe` on Windows (Setup/Portable packages). If it is not on your `PATH`, invoke the full path to the executable. The Linux standalone GUI binary also accepts `--cli` followed by the same arguments.

**Important:** The in-app command field is **not a shell**. Enter **one CLI command at a time**. Do not add terminal-only syntax such as `|`, `>`, `&&`, shell variable expansion, or `cd`. Those belong in a regular terminal. Always quote file and directory names containing spaces.

**Linux terminal examples** (replace executable filename with your actual downloaded file):

```bash
# Packaged command-line executable:
muksmatt-cli --help
muksmatt-cli --version

# Or run the CLI embedded in a Linux standalone GUI executable:
./muKsMaTT-Linux-amd64Standalone --cli --help
```

**Windows PowerShell examples:**

```powershell
# From a folder containing the packaged CLI binary:
.\muksmatt-cli.exe --help
.\muksmatt-cli.exe --version

# Or specify its full executable path:
& "C:\Apps\muKsMaTT\muksmatt-cli.exe" --help
```

When using the **CLI tab**, you can enter `muksmatt-cli --help` on either desktop platform; the app runs the command internally. Unlike a terminal, the CLI tab normally displays the collected command output after the process finishes.

### 2. Basic commands and what they do

| Command | Purpose |
| --- | --- |
| `muksmatt-cli --help` | Show all desktop CLI commands and options. |
| `muksmatt-cli --version` | Show the installed CLI version. |
| `muksmatt-cli tools` | Show the FFmpeg/FFprobe tools muKsMaTT can use (alias: `doctor`). |
| `muksmatt-cli scan SOURCE` | Find DVD titles and report their durations. |
| `muksmatt-cli metadata SOURCE` | Inspect metadata for a selected/longest DVD title. |
| `muksmatt-cli remux SOURCE` | Remux a DVD to MKV without re-encoding compatible streams. |
| `muksmatt-cli bluray scan SOURCE` | List Blu-ray MPLS playlists, durations, and chapters. |
| `muksmatt-cli bluray tracks --playlist ID SOURCE` | Inspect the actual video/audio/subtitle streams in one Blu-ray playlist. |
| `muksmatt-cli bluray remux ...` | Make a Blu-ray MKV or combined M2TS. |
| `muksmatt-cli bluray demux ...` | Extract Blu-ray streams directly into separate files. |
| `muksmatt-cli --batch ...` | Remux a collection of DVD folders/ISOs; this is **not** Blu-ray batch processing. |

`SOURCE` is a placeholder, not a literal filename. DVD sources can be DVD folder structures (including `VIDEO_TS`), supported DVD ISO files, or accessible physical drives. Blu-ray sources can be BDMV disc-root folders, supported Blu-ray ISOs, or accessible devices with configured libbluray support. A filename ending in `.iso` does **not** by itself tell muKsMaTT whether the disc is DVD or Blu-ray: use `bluray` commands for Blu-ray input.

#### Example A: scan a DVD

```bash
muksmatt-cli scan "/media/movies/MyDVD"
```

The result lists DVD **title numbers** and their durations. The `longest` marker identifies the default title. On discs with alternate cuts or unusual structures, inspect the titles and choose the correct one with `--title` rather than assuming the longest is always right.

#### Example B: inspect DVD metadata

```bash
muksmatt-cli metadata --title 1 "/media/movies/MyDVD"
```

This inspects **DVD title 1**. Omit `--title 1` to inspect the longest available title. The output helps you identify tracks and stream indexes before making a selective remux.

#### Example C: remux a DVD to MKV

```bash
muksmatt-cli remux --title 1 --output "/media/output" "/media/movies/MyDVD"
```

This remuxes the selected DVD title into an MKV in the **output directory**, preserving its compatible video/audio/subtitle tracks without a lossy re-encode. DVD `remux --output` takes a **directory**, not an output filename. When `--output` is omitted for a DVD folder or ISO, muKsMaTT chooses the default destination beside the source. For a physical DVD drive, specify `--output` explicitly.

### 3. Basic Blu-ray workflow: scan, inspect, remux

For a Blu-ray folder, select the **disc root** that contains `BDMV` (for example, `/media/blurays/MyMovie`), not an arbitrary `.m2ts` file inside `BDMV/STREAM`.

#### Step 1: list available Blu-ray playlists

```bash
muksmatt-cli bluray scan "/media/blurays/MyMovie"
```

Each entry contains a five-digit **playlist ID**, runtime and chapter count. A `longest` marker may appear, but commercial Blu-rays can include deliberately looping playlists much longer than the real film. **Use the actual intended playlist ID** for further steps. The example `00800` below is illustrative; it is not the correct ID for every disc.

#### Step 2: inspect the selected playlist's tracks

```bash
muksmatt-cli bluray tracks --playlist 00800 "/media/blurays/MyMovie"
```

This displays each stream's **absolute input index**, type, PID, codec, language (when available), and title. For example, a disc might list video index `0`, English DTS-HD MA audio index `1`, and Norwegian PGS subtitles index `9`. The stream **index** is the number used with `--streams`; the hex PID such as `0x1100` is *not* a stream index.

Language detection uses Blu-ray navigation metadata when available. Do not infer an unspecified language from a stream's position.

#### Step 3: remux the selected Blu-ray to MKV

```bash
muksmatt-cli bluray remux --playlist 00800 --output "/media/output/movie.mkv" "/media/blurays/MyMovie"
```

This creates **one MKV** with the playlist's compatible video, audio and subtitle streams and its available chapter markers. Blu-ray `remux --output` takes a **full destination filename** ending in `.mkv` or `.m2ts`, unlike the DVD remux command. The existing output file is not overwritten. For MKV, Blu-ray LPCM audio is converted to **lossless FLAC** when needed; other compatible streams are copied.

### 4. Advanced commands: precise stream and chapter selection

#### Choose DVD title and individual tracks

```bash
# Choose title 2 and only input streams 0, 1, and 3:
muksmatt-cli remux --title 2 --streams 0,1,3 --output "/media/output" "/media/movies/MyDVD"

# Same title and tracks, without DVD chapters:
muksmatt-cli remux --title 2 --streams 0,1,3 --no-chapters --output "/media/output" "/media/movies/MyDVD"
```

`--streams` takes **comma-separated absolute stream indexes**, with no spaces required. Look them up for the selected title before using them; the indexes will differ between discs. `--no-chapters` omits chapter markers from the final MKV. Without `--streams`, normal remux behavior selects all supported media streams.

#### Choose only certain Blu-ray streams

```bash
# Video 0, audio 1, and subtitle 9 (example indexes):
muksmatt-cli bluray remux --playlist 00800 --streams 0,1,9 --output "/media/output/selected.mkv" "/media/blurays/MyMovie"

# Make the same selection without chapters:
muksmatt-cli bluray remux --playlist 00800 --streams 0,1,9 --no-chapters --output "/media/output/no-chapters.mkv" "/media/blurays/MyMovie"
```

The result is a smaller stream selection, not a lower-quality transcode. The source playlist is still read through libbluray rather than treating the first physical M2TS clip as the entire movie.

#### Remux Blu-ray video, audio and PGS subtitles together into M2TS

```bash
# All compatible streams in ONE Blu-ray-style M2TS file:
muksmatt-cli bluray remux --playlist 00800 --output "/media/output/movie.m2ts" "/media/blurays/MyMovie"

# Or choose only particular video/audio/PGS streams:
muksmatt-cli bluray remux --playlist 00800 --streams 0,1,9 --output "/media/output/selected.m2ts" "/media/blurays/MyMovie"
```

**M2TS remux is not DEMUX.** It writes one transport-stream file holding the selected compatible video, audio and PGS subtitle tracks. **Chapters are always excluded**, because this M2TS output does not carry MKV-style chapter metadata. Unsupported formats such as an imported SRT text subtitle are rejected rather than silently lost; use **MKV** for mixed text-subtitle workflows.

Some MediaInfo tools display language tags for M2TS audio but not for PGS subtitles. Even original Blu-ray `BDMV/STREAM/*.m2ts` files commonly omit subtitle language labels: those languages may instead live in the Blu-ray `.mpls`/`.clpi` navigation metadata. Missing PGS language labels alone do not establish missing subtitles.

### 5. Advanced Blu-ray DEMUX: direct elementary-stream extraction

Use `bluray demux` when you want **separate files**, not one MKV or combined M2TS. This command maps selected streams directly from the libbluray playlist to output files **without making an intermediate MKV**.

**Important output rule:** `bluray demux --output` must identify an **existing writable directory**. The CLI creates a new playlist-specific subdirectory there; it does not write directly to a requested movie filename.

#### Demux every media stream and chapters

```bash
muksmatt-cli bluray demux --playlist 00800 --output "/media/output" "/media/blurays/MyMovie"
```

The created subfolder can contain codec-native elementary files such as `track-00.h264`, `track-01.dts`, `track-09.sup`, and `Chapters.txt`. Depending on the original video codec, video might instead use `.h265` or `.mpeg2`; Blu-ray LPCM becomes lossless `.flac` for elementary extraction. Subtitle `.sup` files are PGS; `Chapters.txt` uses the OGM-style chapter text format and is not a video or subtitle stream.

#### Demux just selected audio and subtitles

```bash
# Extract example English audio (1) and Norwegian PGS (9), plus chapters:
muksmatt-cli bluray demux --playlist 00800 --streams 1,9 --output "/media/output" "/media/blurays/MyMovie"

# Same tracks without a chapter text file:
muksmatt-cli bluray demux --playlist 00800 --streams 1,9 --no-chapters --output "/media/output" "/media/blurays/MyMovie"
```

This is useful for testing or repurposing tracks without extracting a large video stream.

#### Export video as elementary H.264/HEVC, video-only M2TS, or both

```bash
# Codec-native elementary video (the default):
muksmatt-cli bluray demux --playlist 00800 --streams 0 --video elementary --output "/media/output" "/media/blurays/MyMovie"

# VIDEO-ONLY M2TS transport stream:
muksmatt-cli bluray demux --playlist 00800 --streams 0 --video m2ts --output "/media/output" "/media/blurays/MyMovie"

# BOTH outputs; can require approximately twice the video storage:
muksmatt-cli bluray demux --playlist 00800 --streams 0 --video both --output "/media/output" "/media/blurays/MyMovie"
```

`--video` changes **only the video export**. Even if audio and subtitles are also selected, `bluray demux --video m2ts` makes a **video-only M2TS**, with selected audio and subtitles written to **separate** files. To put video, audio and PGS subtitles **together** inside one M2TS, use `bluray remux --output movie.m2ts` instead.

### 6. Advanced DVD batch processing

BATCH applies to **DVD** collections (DVD ISO files and folders containing `VIDEO_TS`); it does not automatically batch Blu-ray MPLS titles.

```bash
# Scan a collection and write output beside each DVD source:
muksmatt-cli --batch "/media/DVD-collection"

# Put the completed MKVs under a chosen output root:
muksmatt-cli --batch "/media/DVD-collection" "/media/output"

# Save a persistent activity log as well:
muksmatt-cli --batch --log "/media/output/batch.log" "/media/DVD-collection" "/media/output"
```

Batch mode scans DVD titles and selects the longest readable title for each source, then remuxes compatible streams to MKV. It can continue to subsequent items if one fails. Use `--log` to keep activity information. The destination root and log parent directory must be writable; existing completed output files are not silently overwritten.

### 7. Windows examples

These are **terminal/PowerShell** examples; inside the Windows CLI tab, omit `.\` and type `muksmatt-cli` as shown in the earlier examples.

```powershell
# DVD folder or physical drive:
.\muksmatt-cli.exe scan "D:\"
.\muksmatt-cli.exe remux --title 1 --output "E:\Movies" "D:\"

# Blu-ray BDMV folder:
.\muksmatt-cli.exe bluray scan "D:\"
.\muksmatt-cli.exe bluray tracks --playlist 00800 "D:\"

# One MKV with chapters or one M2TS without chapters:
.\muksmatt-cli.exe bluray remux --playlist 00800 --output "E:\Movies\movie.mkv" "D:\"
.\muksmatt-cli.exe bluray remux --playlist 00800 --output "E:\Movies\movie.m2ts" "D:\"

# Direct stream extraction into an existing folder:
.\muksmatt-cli.exe bluray demux --playlist 00800 --streams 1,9 --output "E:\Movies" "D:\"
```

Here, `D:\` represents an **accessible** disc location and `E:\Movies` a writable destination; replace them with your own paths. The DVD and Blu-ray commands are distinct. Real physical drives and copy-protected media depend on OS device access and separate, legally configured AACS/BD+ or DVD decryption support; such access is not guaranteed by an installed binary.

### 8. Verifying your output

FFprobe and MediaInfo are useful for inspecting results after a job finishes. These are **terminal commands**, not built-in muKsMaTT CLI subcommands:

```bash
# List audio, video, subtitle streams and their metadata:
ffprobe -v error -show_entries stream=index,codec_type,codec_name:stream_tags=language,title -of json "/media/output/movie.mkv"

# Check MKV chapter counts:
ffprobe -v error -show_entries chapter=start_time,end_time:chapter_tags=title -of json "/media/output/movie.mkv"

# Inspect a combined M2TS (expect no MKV-style chapters):
ffprobe -v error -show_streams -show_chapters "/media/output/movie.m2ts"

# Alternative human-readable report, if MediaInfo is installed:
mediainfo "/media/output/movie.m2ts"
```

For direct DEMUX, inspect the new playlist-specific subfolder and check that selected files are nonempty. For an MKV, verify the intended audio language, subtitle language, chapter names and playback synchronization. For an M2TS, verify the **actual presence** of audio/PGS tracks rather than assuming missing PGS language labels mean the streams disappeared.

### 9. Common mistakes and troubleshooting

| Symptom | What to check |
| --- | --- |
| The selected Blu-ray playlist shows an implausible 15-hour runtime | The longest playlist may be a looping decoy. Use `bluray scan`, then explicitly choose the intended `--playlist`. |
| `bluray tracks` or `bluray demux` reports missing `--playlist` | These commands require an explicit playlist number. |
| The wrong audio/subtitle streams were extracted | Use `bluray tracks --playlist ID`, then specify **stream indexes**, not PIDs, with `--streams`. |
| The DVD remux command rejects the destination | DVD `remux --output` is an **output directory**. Blu-ray `remux --output` is a **new .mkv/.m2ts filename**. |
| Blu-ray DEMUX rejects the output folder | Create an existing, writable directory before using `--output`. |
| M2TS fails with SRT, ASS or another codec | M2TS supports only appropriate Blu-ray-compatible codecs. Use MKV for text subtitles, or convert to a supported PGS source before muxing. |
| The M2TS has no chapters | Expected: M2TS output intentionally excludes chapters. Use MKV to retain them. |
| M2TS PGS tracks have no displayed language | Not necessarily an error: commercial Blu-rays commonly store subtitle language information in playlist/clip navigation metadata. |
| A Blu-ray ISO or physical drive cannot be read | Verify actual disc structure, file/device permissions, libbluray-enabled FFmpeg/FFprobe and required authorized protection support. A file extension alone is not enough. |
| Running a command in the CLI tab doesn't understand `|` or `&&` | The field is not a shell. Use a terminal to chain commands or redirect output. |
| A command reports that the output file already exists | Choose a new destination filename. Do not rely on an automatic overwrite. |

**Cancellation and storage:** The CLI tab has **Cancel**; in a terminal, use the normal interrupt (for example, **Ctrl+C**). Processing large Blu-rays can take time, and exporting both elementary video and video-only M2TS consumes additional disk space. Ensure the output location has adequate free space and inspect the completed files before deleting anything from the source.

### 10. Quick command reference

```text
# General
muksmatt-cli --help
muksmatt-cli --version
muksmatt-cli tools

# DVD
muksmatt-cli scan SOURCE
muksmatt-cli metadata [--title N] SOURCE
muksmatt-cli remux [--title N] [--streams 0,1,2] [--no-chapters] [--output DIR] SOURCE
muksmatt-cli --batch [--log FILE] MOVIES_ROOT [OUTPUT_ROOT]

# Blu-ray
muksmatt-cli bluray scan SOURCE
muksmatt-cli bluray tracks --playlist 00800 SOURCE
muksmatt-cli bluray remux --playlist 00800 --output movie.mkv [--streams 0,1,9] [--no-chapters] SOURCE
muksmatt-cli bluray remux --playlist 00800 --output movie.m2ts [--streams 0,1,9] SOURCE
muksmatt-cli bluray demux --playlist 00800 --output EXISTING_DIR [--streams 0,1,9] [--no-chapters] [--video elementary|m2ts|both] SOURCE
```

Remember: this guide describes the **Windows/Linux desktop CLI**. Advanced GUI **MUX** can additionally mix imported files and edit stream language/title metadata; those particular editing controls are not a separate desktop CLI command.
