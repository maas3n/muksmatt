# Blu-ray Phase 4 — Windows and Linux native runtime packaging (development)

This PR continues from PR #13. It does not modify DVD remux, DVD decryption,
Android JNI, Android signing or the existing user-interface layout.

## Windows packaging

The Windows native libbluray navigator is now compiled with MSYS2 MinGW64 and
the installed libbluray headers and library. The build also installs MinGW64
libaacs and libbdplus runtime dependencies; no keys or authorization data are
bundled.

A deterministic Windows DLL closure walker reads each DLL import table via
`objdump -p`. It refuses missing dependencies, stages the navigator and
required DLLs, records their SHA-256 hashes, and copies the three Blu-ray core
license notices. It deliberately excludes Windows system DLLs.

The Windows CI runs the DLL closure unit tests, builds the real navigator, and
**launches the staged binary from an isolated directory**. The disabled unified
release builder stages these files in the installer and portable distribution.
Additional DLLs are copied alongside private FFmpeg because its built-in
libbluray may load libaacs and libbdplus dynamically.

Windows remux depends on the bundled FFmpeg/FFprobe exposing `bluray:`.
The release workflow fails closed if either build lacks it. Windows distribution
has not been published or installed on physical hardware.

## Linux packaging

The Linux portable archive already carries the navigator. Phase 4 additionally
embeds the native navigator in the single-file self-extracting standalone
binary, preserves execute permissions, and routes `bluray` CLI subcommands
from the launcher. Its self-test checks the extracted native helper.

The standalone bundler now includes `libaacs.so.0`, `libbdplus.so.0` and
`libudfread.so.0` that libbluray may load dynamically, including their
transitive dependencies and Debian copyright notices. The Linux .deb declares
the AACS/BD+ runtime dependencies explicitly. No system multimedia tools are
overwritten.

Linux release and CI now require both FFmpeg and FFprobe to advertise
`bluray:`; otherwise packaging fails rather than shipping non-working
Blu-ray commands.

## Existing native behavior

The existing opt-in commands remain:

    muksmatt-cli bluray scan BDMV_ROOT | DISC.iso | /dev/sr0
    muksmatt-cli bluray remux --output output.mkv [--playlist 00800] SOURCE

The selected Blu-ray playlist is remuxed directly via the libbluray protocol,
with user-selected stream mapping and MPLS chapters. Compatible video/audio/
subtitle tracks are copied and `pcm_bluray` is always losslessly converted to
FLAC. The DVD input policy `100M / 100M / +genpts` remains DVD-only.

## Validation boundaries

CI compilation and a native helper `--version` test are **not** evidence
that an encrypted Blu-ray has decrypted or that Windows/Linux devices can
read protected media successfully. Library availability does not grant access
to disc-specific keys, certificates or BD+ data. Multi-angle titles, genuine
BDMV/ISO remux duration/timestamp accuracy, encrypted disc variants, physical
drives and AACS/BD+ handling still need real-media tests.

Android/ChromeOS native Blu-ray video reading is a separate follow-on PR,
and remains incomplete. No automatic release is enabled. No user signing
secrets or decryption keys are modified or uploaded.

References: `tools/bluray/bluray_nav.c`,
`packaging/windows/stage-bluray-windows.py`, and
`packaging/linux/build-linux-standalone.sh`.
