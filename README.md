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
