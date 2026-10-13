> **Release status (13 October 2026):** [muKsMaTT v0.1.0-beta.1](https://github.com/maas3n/muksmatt/releases/tag/v0.1.0-beta.1) is a **public GitHub prerelease**, not a stable or Google Play release. Publication is manual and opt-in: pushes and tags never publish automatically. A numbered release requires repository variable `MUKSMATT_RELEASE_ENABLED=true` **and** a manual Unified release run with `confirm_publish=true`. Restore the variable to `false` or unset it after an authorized publication. `MATTRIP_*` Gradle parameters are inherited *internal* compatibility contracts, not the public app identity; never reuse MattRip signing keys.

# Releasing muKsMaTT

muKsMaTT uses one long-lived source branch, `main`, and one product version namespace across Windows, Linux, and Android/ChromeOS.

## Unified release model

Every new public release uses exactly one product tag and one GitHub Release:

- stable: `vMAJOR.MINOR.PATCH`
- preview: `vMAJOR.MINOR.PATCH-alpha.N`, `-beta.N`, or `-rc.N`

Do not create new platform-specific version tags such as `-linux`, `-chromeos`, `-windows`, or separate `devN` release lines. Historical platform-specific tags remain valid historical pointers and are not rewritten.

A unified release contains the platform assets that are ready from the same tagged commit. Typical assets are:

- `muKsMaTT-<version>-Windows-Setup.exe`
- `muKsMaTT-<version>-Windows-All-in-One.exe`
- `muKsMaTT-<version>-Windows-Portable.zip`
- `muKsMaTT-<version>-Linux-amd64.deb`
- `muKsMaTT-<version>-Linux-amd64.tar.gz`
- `muKsMaTT-<version>-Linux-amd64Standalone`
- `muKsMaTT-<version>-Source.tar.gz`
- `muKsMaTT-<version>-Android.apk` — universal APK for Android phones/tablets and Chromebooks with Android app support
- third-party source/provenance/license files
- per-platform checksum manifests
- one combined `SHA256SUMS.txt`

`muKsMaTT-<version>-Android.apk` is the single persistently signed universal APK for both Android phones/tablets and Chromebooks with Android app support.

GitHub also exposes source ZIP/tar archives automatically for the release tag.

## Release principles

1. **Published release tags are immutable.** Never force-move an existing version tag.
2. **Published release assets are immutable.** Fix a released problem in a new version.
3. **All platform payloads come from the same tag/commit.** Windows, Linux, and Android/ChromeOS must not publish different source commits under the same product version.
4. **One tag creates one GitHub Release.** Platform workflows may build independently, but `.github/workflows/release.yml` is the only workflow that publishes GitHub Releases.
5. Generate and verify SHA-256 checksums for release payloads and verify bundled runtime dependencies before publishing.
6. Keep historical development provenance in Git history; obsolete platform-specific public release entries may remain retired after the unified-release cleanup.

## Before tagging

- Merge the intended source into `main`.
- Confirm Windows, Linux, and Android/ChromeOS CI is green.
- For releases containing Android, confirm **Android signing self-test** is green and both pinned signing fingerprints match the intended identities.
- Confirm pinned third-party versions/checksums and licensing/provenance documentation are current.
- Choose a muKsMaTT version/tag that is unused in this fork. Existing MattRip or MattMux tags belong to upstream provenance and must not be reused for muKsMaTT releases.
- Decide whether the release is stable or a shared preview.
- Do not reuse a tag that already exists or already has a GitHub Release.

## Publishing

Create the tag from the exact `main` commit to publish and push it:

```bash
git switch main
git pull --ff-only
git tag vX.Y.Z
git push origin vX.Y.Z
```

For a preview:

```bash
git tag vX.Y.Z-alpha.1
git push origin vX.Y.Z-alpha.1
```

Pushing a tag **does not** start the release workflow. Public publication requires a separately authorized manual run. For the first authorized release, explicitly set the repository Actions variable `MUKSMATT_RELEASE_ENABLED=true`, dispatch **Unified release** on `main`, enter the existing tag (or leave empty for main downloads), and select **confirm_publish**. Without both the variable and confirmation, all build/publish jobs are skipped. Do **not** set that variable during signing setup.

Once manually approved, the **Unified release** workflow:

1. validates the unified tag format;
2. derives one product version plus the Android `versionCode`;
3. builds the Windows payload;
4. builds the Linux payload;
5. builds one signed universal `muKsMaTT-<version>-Android.apk` for both Android and ChromeOS;
6. verifies each platform payload;
7. downloads all platform artifacts into one release job;
8. creates a combined `SHA256SUMS.txt`; and
9. publishes one GitHub Release for that tag.

The workflow refuses to overwrite an existing GitHub Release. Numbered releases are staged as private drafts, uploaded and checksum-verified before becoming visible; partial upload failures leave an unpublished draft rather than a broken public release. A separately dispatched `build_only=true` rehearsal builds and validates packages without running the publisher or requiring release authorization.

Manual dispatch uses an **existing** unified tag or an empty tag for the rolling main snapshot. Manual dispatch does not invent or move tags; it cannot publish without the separate repository opt-in and explicit `confirm_publish` checkbox. The independent **Android signing self-test** may run while publication remains locked.

### First public beta: v0.1.0-beta.1 (completed historical checklist)

**This checklist documents the release process already used to publish the first beta on 13 October 2026. Do not recreate, retag, edit or republish `v0.1.0-beta.1`.** The tag is protected by the active “Immutable release tags” ruleset, and the numbered release assets must remain unchanged. For a future release, choose a **new unused tag**, confirm CI and signing, use `build_only=true` for rehearsals and reserve the two publication approvals for an expressly authorized run.

1. Merge reviewed beta release notes from `docs/releases/v0.1.0-beta.1.md` into `main` first. The metadata step refuses prereleases without a nonempty reviewed notes file at the **tagged commit**.
2. Confirm post-merge Windows, Linux, Android/ChromeOS and release-tool CI all succeeded, the Android signing self-test is green, and GitHub environment secrets/certificate pins are still configured. Check any required `android-release` environment approval rules before dispatching.
3. **Rehearse release packaging without publishing:** In Actions → **Unified release** → Run workflow on `main`, leave `tag` blank, leave `confirm_publish` **unchecked**, and check `build_only`. This mode does the actual Windows, Linux and signed Android release builds and Android startup validation, uploads private Actions build artifacts, then **skips the entire GitHub Release publisher**. It never creates a tag or GitHub Release. Keep `MUKSMATT_RELEASE_ENABLED` absent; rerun/fix until every package-building job passes. This checks the development-snapshot packaging path before the tagged release.
4. The first beta tag was created at the tested commit `f1f91670d2c42416760a43723f37c58ba33cb477`. It now exists and is immutable. Future versions must use a new, previously unused release tag.
5. Confirm the repository GitHub Actions token can create releases (the publish job explicitly requests `contents: write`) and GitHub Actions is enabled. Only the release workflow can publish; ordinary CI cannot.
6. The first beta received explicit publication approval. For **future, separately authorized** releases, set `MUKSMATT_RELEASE_ENABLED=true`, dispatch **Unified release** on `main`, enter that *new* existing tag (a blank field selects a rolling main build), then select `confirm_publish`. Ensure metadata resolves `mode=release` and the intended prerelease setting.
7. The workflow builds every platform from the same tag, verifies Android APK signing, requires Android startup checks, checks the complete asset set and checksums, then stages the numbered release **as a private draft**. Uploaded assets are verified against local SHA-256 digests before the draft becomes public. If an upload/verification fails, it stays private; never publish or overwrite incomplete assets manually. Inspect a failed draft before deciding whether to delete and rerun.
8. Verify the public release is labeled **Pre-release**, not **Latest**, and includes Windows Setup / All-in-One / Portable, Linux Debian/tarball/Standalone, the signed Android/ChromeOS APK, combined SHA256SUMS, notices and all source/provenance packages. Confirm the body includes the beta tester instructions and limitation warnings.
9. After each authorized publication, set `MUKSMATT_RELEASE_ENABLED` to `false` or remove it to lock future publishing. Disabling the variable does **not** cancel an already-authorized run; verify the run completes. Preserve muKsMaTT's app-signing key for future Android upgrades.

Do not accidentally dispatch with a blank tag; it selects **main-downloads**, not a numbered version. Do not enable publication merely to rehearse the workflow. Genuine Blu-ray drive testing and protected-disc coverage remain outstanding and must be labeled experimental.

## Android signing gate

Android uses two deliberately separate signing identities. See [`android/SIGNING.md`](android/SIGNING.md) for the complete setup.

- Public GitHub APKs use the **muKsMaTT app-signing key** from the protected `android-release` environment.
- Play AAB uploads use the separate **muKsMaTT upload key** from the protected `android-play` environment.
- Repository Actions variables `ANDROID_APP_SIGNING_CERT_SHA256` and `ANDROID_UPLOAD_CERT_SHA256` pin the public certificates.
- Run **Android signing self-test** successfully before tagging any release that is expected to include Android.
- The two certificate fingerprints must be different.
- The unified release workflow passes `MATTRIP_REQUIRE_SIGNING=true` and verifies the final APK certificate. It must never fall back to an unsigned or debug-signed production APK.
- The Play workflow verifies the final AAB against the pinned upload certificate.

The original MattRip 1.5.0 was published in the upstream project. The first signed muKsMaTT GitHub APK is published as part of `v0.1.0-beta.1` and uses a **different** application/signing identity. Preserve the existing muKsMaTT signing keys for updates and never reuse an upstream tag.

## Android / ChromeOS versionCode

The unified workflow derives a monotonically ordered Android versionCode from the product version:

- alpha builds sort before beta builds;
- beta builds sort before release candidates;
- release candidates sort before the stable release;
- the next patch/minor/major version sorts after the previous stable release.

Android/ChromeOS purchases remain disabled until production device validation, signing, and purchase-verification readiness are complete. The separate Play bundle workflow is distribution tooling; it does not create GitHub Releases.

muKsMaTT uses application ID `io.github.maas3n.muksmatt`; it is a different Android app from MattMux and MattRip, not an in-place upgrade path for either. Its first public beta APK was signed under muKsMaTT's own persistent release identity. Verify and preserve that identity for all subsequent releases.

## Historical releases

The repository was forked from the MattRip beta-build-6-1 baseline, which originated from MattMux 1.4.19. Any inherited MattMux tags, platform-specific tags, and old release-line history are pre-fork provenance, not muKsMaTT releases. Leave them immutable and do not reuse them for muKsMaTT. New muKsMaTT releases use only new, unused unified tags.

## Emergency fixes

If a published unified release is defective, leave its tag and assets unchanged, fix the problem on `main`, and publish the next unused muKsMaTT product version. Never rebuild an old release in place.
