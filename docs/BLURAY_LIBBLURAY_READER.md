# Blu-ray libbluray stream-reading milestone

This PR introduces the **first working desktop stream-reading/remux execution path** as an opt-in CLI command, separately from the existing DVD and MKV paths. It builds on the pure-Go Blu-ray playlist/chapter parser from PR #11.

## Windows and Linux commands

With an **accessible disc-root folder** containing BDMV/PLAYLIST, and an FFmpeg/FFprobe pair compiled with libbluray:

```text
muksmatt-cli bluray scan /path/to/disc-root
muksmatt-cli bluray remux --output movie.mkv /path/to/disc-root
muksmatt-cli bluray remux --playlist 00800 --output movie.mkv --streams 0,1,2 /path/to/disc-root
muksmatt-cli bluray remux --no-chapters --output movie.mkv /path/to/disc-root
```

Windows paths may be used in place of the Linux path. An already-mounted optical disc with a BDMV structure can be accessed by its mounted folder. The program does not itself mount devices.

The read path uses FFmpeg's real `bluray:` protocol, passing the selected MPLS playlist number. It checks FFmpeg and FFprobe for `bluray` protocol availability, probes actual A/V/subtitle stream indexes, maps them explicitly, and then remuxes **directly from the selected playlist** into a new Matroska file. No temporary MKV is built before remuxing.

- Video, PGS subtitles and compatible audio use `-c copy`.
- **Every** selected `pcm_bluray` audio stream uses lossless **FLAC**, indexed by its output audio position.
- Chapters are reconstructed from MPLS `PlayListMark` entries via a temporary FFmetadata input with 45 kHz timestamps, then mapped into Matroska.
- Final output is published only after FFprobe stream verification and atomic no-overwrite finalization. Failed runs leave no published final MKV.
- Existing DVD-only large probe and `+genpts` flags are not applied to this source.

## Protection and supported media

FFmpeg's libbluray protocol can use **locally configured** libaacs and libbdplus libraries and their existing authorized metadata/key material to read supported protected titles. muKsMaTT does not download or distribute keys, AACS certificates or BD+ VM data and **does not claim all encrypted discs will open**. The new source path rejects absent protocol support and reports probable decryption or media access failures.

**Current limitations:** disc-root folders (including mounted discs) only. Standalone ISO files and raw optical devices have not yet been connected to the libbluray reader. Menus, playlist obfuscation resolution, seamless branching verification, multi-angle selection, physical-disc testing and advanced selection through the GUI are pending. This PR does **not** enable those features.

## Android and ChromeOS

Android uses Kotlin/JNI and the Storage Access Framework (SAF), not the Go desktop backend. A bounded Kotlin MPLS PlayList/chapters parser and JUnit tests are included for Android/ChromeOS metadata parity. No Android libbluray reader or Blu-ray UI action is implemented in this PR. The next Android-specific milestone should define a bounded SAF read/seek adapter and native libbluray integration, followed by the same chapter and LPCM→FLAC policy and emulator/device tests. Do **not** route Blu-ray content through the existing DVD native engine or assume that the ChromeOS file provider exposes raw optical drives.

## Validation

Pure Go unit tests exercise protocol validation, explicit playlist options, absence of DVD-only flags, A/V/subtitle map generation, mixed TrueHD/LPCM/AC-3 FLAC assignment, exact MPLS chapter tick metadata and invalid stream selections.

GitHub Windows and Linux CI compile and run the code. Successful CI does not establish that a particular user disc or AACS/BD+ setup has passed decryption, or that the bundled FFmpeg builds support libbluray. Physical media and actual libbluray read/remux tests are still required before public release.

Unified production publishing remains disabled.
