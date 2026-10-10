> **muKsMaTT development:** The Unified release workflow is locked against automatic publishing: pushes and tags never launch it. The repository variable `MUKSMATT_RELEASE_ENABLED` must explicitly equal `true`, and the manual run must have `confirm_publish=true`. Leave that variable **unset** while signing is being configured. No muKsMaTT stable or Play Store release has been published. The signing environment names and `MATTRIP_*` Gradle parameters are temporary inherited internal compatibility contracts; do not reuse MattRip's private keys.

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

### First public beta: v0.1.0-beta.1 readiness checklist

**Preparation is not publication.** As long as `MUKSMATT_RELEASE_ENABLED` is absent or not `true`, a normal Unified release dispatch skips its entire build/publish chain. The dedicated `build_only` option intentionally permits signed package validation while always skipping public publishing. Do not run with both publication approvals merely to test the pipeline.

1. Merge reviewed beta release notes from `docs/releases/v0.1.0-beta.1.md` into `main` first. The metadata step refuses prereleases without a nonempty reviewed notes file at the **tagged commit**.
2. Confirm post-merge Windows, Linux, Android/ChromeOS and release-tool CI all succeeded, the Android signing self-test is green, and GitHub environment secrets/certificate pins are still configured. Check any required `android-release` environment approval rules before dispatching.
3. **Rehearse release packaging without publishing:** In Actions → **Unified release** → Run workflow on `main`, leave `tag` blank, leave `confirm_publish` **unchecked**, and check `build_only`. This mode does the actual Windows, Linux and signed Android release builds and Android startup validation, uploads private Actions build artifacts, then **skips the entire GitHub Release publisher**. It never creates a tag or GitHub Release. Keep `MUKSMATT_RELEASE_ENABLED` absent; rerun/fix until every package-building job passes. This checks the development-snapshot packaging path before the tagged release.
4. Confirm `v0.1.0-beta.1` is absent from Git tags and GitHub Releases. Make the immutable tag on that exact tested `main` commit; record its SHA. Do not tag before these preparation commits are merged.
5. Confirm the repository GitHub Actions token can create releases (the publish job explicitly requests `contents: write`) and GitHub Actions is enabled. Only the release workflow can publish; ordinary CI cannot.
6. **Stop here until explicit public publication approval.** At authorization time, set `MUKSMATT_RELEASE_ENABLED=true` as a repository Actions variable, dispatch **Unified release** on `main`, enter the exact tag `v0.1.0-beta.1` (do **not** leave the field blank: that selects a rolling main build), then explicitly select `confirm_publish`. Check the metadata resolved mode is `release` and `prerelease=true`.
7. The workflow builds every platform from the same tag, verifies Android APK signing, requires Android startup checks, checks the complete asset set and checksums, then stages the numbered release **as a private draft**. Uploaded assets are verified against local SHA-256 digests before the draft becomes public. If an upload/verification fails, it stays private; never publish or overwrite incomplete assets manually. Inspect a failed draft before deciding whether to delete and rerun.
8. Verify the public release is labeled **Pre-release**, not **Latest**, and includes Windows Setup / All-in-One / Portable, Linux Debian/tarball/Standalone, the signed Android/ChromeOS APK, combined SHA256SUMS, notices and all source/provenance packages. Confirm the body includes the beta tester instructions and limitation warnings.
9. Remove the temporary `MUKSMATT_RELEASE_ENABLED` repository Actions variable after the run so publishing is locked again. Removing the variable does **not** cancel a publication run already authorized and started; monitor the run until it completes. Preserve the same Android app-signing key for later updates.

Do not accidentally dispatch with a blank tag; it selects **main-downloads**, not `v0.1.0-beta.1`. Do not set the variable or publish anything just to rehearse the workflow. Physical Blu-ray drive testing and protected-disc coverage remain outstanding and must be labeled experimental.

## Android signing gate

Android uses two deliberately separate signing identities. See [`android/SIGNING.md`](android/SIGNING.md) for the complete setup.

- Public GitHub APKs use the **muKsMaTT app-signing key** from the protected `android-release` environment.
- Play AAB uploads use the separate **muKsMaTT upload key** from the protected `android-play` environment.
- Repository Actions variables `ANDROID_APP_SIGNING_CERT_SHA256` and `ANDROID_UPLOAD_CERT_SHA256` pin the public certificates.
- Run **Android signing self-test** successfully before tagging any release that is expected to include Android.
- The two certificate fingerprints must be different.
- The unified release workflow passes `MATTRIP_REQUIRE_SIGNING=true` and verifies the final APK certificate. It must never fall back to an unsigned or debug-signed production APK.
- The Play workflow verifies the final AAB against the pinned upload certificate.

The original MattRip 1.5.0 was published in the upstream project; muKsMaTT has no numbered release yet. Configure dedicated muKsMaTT Android signing identities before its first published APK, and never reuse an upstream tag.

## Android / ChromeOS versionCode

The unified workflow derives a monotonically ordered Android versionCode from the product version:

- alpha builds sort before beta builds;
- beta builds sort before release candidates;
- release candidates sort before the stable release;
- the next patch/minor/major version sorts after the previous stable release.

Android/ChromeOS purchases remain disabled until production device validation, signing, and purchase-verification readiness are complete. The separate Play bundle workflow is distribution tooling; it does not create GitHub Releases.

muKsMaTT uses the separate application ID `io.github.maas3n.muksmatt`, so it is a different Android app from MattMux and is not an in-place upgrade path for MattMux APKs. Establish and preserve muKsMaTT's own release-signing identity before the first public APK, and verify its certificate before publishing.

## Historical releases

The repository was forked from the MattRip beta-build-6-1 baseline, which originated from MattMux 1.4.19. Any inherited MattMux tags, platform-specific tags, and old release-line history are pre-fork provenance, not muKsMaTT releases. Leave them immutable and do not reuse them for muKsMaTT. New muKsMaTT releases use only new, unused unified tags.

## Emergency fixes

If a published unified release is defective, leave its tag and assets unchanged, fix the problem on `main`, and publish the next unused muKsMaTT product version. Never rebuild an old release in place.
