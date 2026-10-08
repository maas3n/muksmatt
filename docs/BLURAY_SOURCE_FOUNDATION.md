# Blu-ray source foundation (development only)

This is the first additive Blu-ray milestone in muKsMaTT. It deliberately leaves the current DVD, MKV, REMUX/DEMUX, ADVANCED, BATCH, CLI, Android native processing, UI layout and release packaging unchanged.

## Implemented on Windows and Linux

- Identify a disc-root folder containing BDMV/PLAYLIST or an explicitly selected BDMV folder. Folder-name and .mpls extension casing are tolerated.
- Recognize .iso as a **candidate** only. Do not claim an ISO is a verified Blu-ray based on its filename.
- Enumerate numbered .mpls navigation files for disc-root folders.
- Parse main-angle PlayItem IN/OUT intervals into a continuous 45 kHz playlist timeline and type-1 PlayListMark chapters. Handle multi-clip clock offsets without manufacturing chapter points.
- Select a playlist by explicit five-digit ID, or choose the longest duration, using lowest-numbered ID to break ties.
- Reject corrupt/oversized MPLS files, out-of-bounds section lengths, invalid item timing, duplicate playlist IDs and excessively long durations.
- Use pure Go, with no new library or Bash dependency.

Run Windows/Linux Go tests in the existing CI. Unit fixtures include a two-clip discontinuous timeline, explicit vs. longest playlist, chapter deduplication and malformed navigation structures.

## Not yet implemented

This milestone **does not** open or demux Blu-ray media, read encrypted streams, read inside ISO images, navigate optical drives, select video/audio/subtitle tracks, convert LPCM to FLAC, produce MKVs or change any existing GUI behavior. It does not add a functional Blu-ray button.

Next stages should connect this metadata to a libbluray-backed reader for folders, ISO and optical drives (with libaacs/libbdplus as permitted and configured); then add source stream probing and per-track preservation. The standing audio rule from BDRemux-Decrypt.sh is **pcm_bluray to FLAC**, with other compatible audio codecs copied.

Any platform-specific native decryption and Android integration must receive separate CI and physical-disc validation. No signing keys or decryption material are shipped or fetched by this module.

Reference: the user-provided BDRemux-Decrypt.sh playlist/chapter design; the script is **not** being bundled or executed.
