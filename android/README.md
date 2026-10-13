# muKsMaTT for Android / ChromeOS

This directory contains muKsMaTT's **experimental** Android/ChromeOS application. The Android application ID is **`io.github.maas3n.muksmatt`**. It is separate from MattRip and MattMux: installing muKsMaTT does not replace or upgrade either application.

## Public beta and supported devices

The first signed GitHub prerelease is [muKsMaTT v0.1.0-beta.1](https://github.com/maas3n/muksmatt/releases/tag/v0.1.0-beta.1). Download the [universal Android APK](https://github.com/maas3n/muksmatt/releases/download/v0.1.0-beta.1/muKsMaTT-0.1.0-beta.1-Android.apk) and verify it using the release's [SHA256SUMS.txt](https://github.com/maas3n/muksmatt/releases/download/v0.1.0-beta.1/SHA256SUMS.txt).

- Android **8.0 (API 26)+**; supported native ABIs: **arm64-v8a** and **x86_64**.
- Chromebooks need the Android-app runtime; there is no separate ChromeOS APK.
- GitHub prerelease APKs use the dedicated, persistent **muKsMaTT app-signing key**. Do not reuse MattRip credentials or expect in-place updates from MattRip/MattMux.
- This is **not a Google Play release**. Production Play purchases and public Play rollout remain disabled/pending separate validation.
- The beta is intended for community testing; real phones, Chromebooks, physical discs and protected commercial Blu-rays still require independent validation.

For the user-facing DVD, Blu-ray, installation, and testing overview, see the [repository README](../README.md) and the [beta testing guide](../docs/releases/v0.1.0-beta.1-testing.md).

## DVD and container workflows

The Android application provides:

- **REMUX/DEMUX:** user-selected `VIDEO_TS` DVD folders, read-only UDF DVD ISOs, and MKV sources; longest-title DVD discovery, stream selection, optional chapters, direct remux/demux and progress/cancellation.
- **ADVANCED:** combine selected compatible media streams into MKV and extract selected streams using its DEMUX action. Subtitle input may require matching `.idx` plus `.sub` files.
- **BATCH REMUX:** process collections of DVD folders and ISO sources through Android's Storage Access Framework (SAF).
- **In-app CLI:** DVD-oriented `scan`, `metadata`, `remux` and `--batch` commands with selected streams and chapter control.

Input/output use the system's SAF file and folder pickers rather than assuming ordinary filesystem paths. ISO reads require a **seekable** provider. Some operations stage media in app-private storage and require free space; output may require seek and rename support. Failed or cancelled output should not be treated as a completed MKV.

### DVD and CSS behavior

Native DVD handling uses FFmpeg/libav, libdvdnav/libdvdread for title discovery, libudfread for ISO access, and an optional libdvdcss sector-reading path in **direct GitHub/development builds**. The CSS-enabled path remains part of the DVD engine, rather than a second remuxer. The **Google Play build intentionally disables CSS** pending a separate distribution review. DVD-specific analysis/generation policy is not applied to non-DVD MKV or Blu-ray sources.

CSS support for DVD folder/ISO input does **not** imply that a phone can read a USB optical drive. Android USB CD/DVD/Blu-ray transport has not been established as supported.

## Blu-ray — experimental native remux

An independently developed native `libbluray` reader accepts user-granted **BDMV folder trees** and **seekable Blu-ray ISO documents** through SAF. It can read a chosen or longest playlist and remux selected compatible streams into Matroska with chapters, converting selected Blu-ray LPCM (`pcm_bluray`) to **lossless FLAC**. Self-authored unencrypted fixtures have been tested in Android emulators, including API 26 and API 35 (4 KB/16 KB pages).

**Important scope:** This describes the native capability and test path, not a promise that the existing DVD **REMUX/DEMUX** or **ADVANCED** GUI provides finished Blu-ray actions. Android Blu-ray GUI integration, real-device compatibility, external USB Blu-ray drives, and AACS/BD+ protected-disc decryption are **not validated**. A green emulator run cannot establish that a protected retail disc can be decrypted or read.

See [Blu-ray real-media validation](../docs/BLURAY_REAL_MEDIA_VALIDATION.md) and [Android native remux work](../docs/BLURAY_PHASE5_ANDROID_REMUX.md).

## Signing, Play and billing

muKsMaTT uses separate Android identities for the direct-download APK and any future Google Play AAB upload. The repository's protected `android-release` environment holds the app-signing credentials; `android-play` uses a distinct Play-upload identity. CI rejects absent or mismatched production credentials and verifies the certificate fingerprint of signed outputs.

See [SIGNING.md](SIGNING.md) for certificate and workflow requirements and [PLAY_CONSOLE.md](PLAY_CONSOLE.md) for future Play preparation. `BuildConfig.ENABLE_BILLING_PURCHASES` is currently **false**; an inherited Billing dependency does not mean that purchases can be made or that Play distribution has launched.

Never commit keystores, credentials or signing passwords.

## Native runtime and licensing

The bundled Android FFmpeg **9.0.1** runtime uses an **LGPL-only** configuration (GPL/nonfree disabled). Native DVD/ISO and Blu-ray libraries are integrated separately where appropriate. Libraries are built and packaged for both supported 64-bit ABIs and must satisfy the Android 16 KB-page alignment checks.

The native DVD path uses `libmattmux_jni.so` as an inherited internal library name; this does **not** change the public muKsMaTT app identity. Native build requirements, pinned sources, licenses, and configuration are in [native/README.md](native/README.md), [native/FFMPEG_LGPL_POLICY.md](native/FFMPEG_LGPL_POLICY.md) and [THIRD_PARTY.md](../THIRD_PARTY.md).

## Development build

CI uses JDK 17, Gradle 9.6.0, Android SDK API 36 and Android NDK `30.0.16248370`, in addition to the native build prerequisites.

From the repository root:

```bash
export ANDROID_NDK_HOME="$ANDROID_HOME/ndk/30.0.16248370"
bash android/native/build-ffmpeg-android.sh
gradle -p android :app:testDebugUnitTest :app:lintDebug :app:assembleDebug
```

This produces **development artifacts**, not the signed public GitHub prerelease. Signed releases must go through the authorized [Unified release workflow](../RELEASING.md). Do not manually sign a debug APK with a production key.
