# Signed Android release startup validation

Rehearsal 38079000465 built all three platform packages successfully, but the
Android startup matrix required BluraySmokeInstrumentation from `src/debug`.
The production APK correctly excluded that instrumentation and its fixtures.

The release workflow now builds `assembleReleaseAndroidTest` alongside the
unchanged production release APK. The separate test-only APK uses the same
signing identity, targets `io.github.maas3n.muksmatt`, and carries the authored
Blu-ray ISO/BDMV assets and DocumentsProvider. The shared smoke test exercises
the installed application's native engine through real cross-UID SAF grants.
Fixture preparation is limited to callers with the same signing certificate.
The instrumentation launches the companion bootstrap after Android's target
process restart, which clears transient URI grants. Only that bootstrap launch
uses UiAutomation's shell command; the test does not adopt shell permissions.
Provider calls and media operations run under the production application's UID.

`verify-startup-apks.py` checks matching certificates, instrumentation target,
fixture presence and isolation from the production APK before emulator launch.
The reusable startup workflow explicitly installs both APKs and requires the
Blu-ray pass marker plus a successful instrumentation result on API 26, API 35
4 KB and API 35 16 KB. Existing debug DVD/demux startup coverage remains enabled.
Pull requests additionally exercise release source sets using a disposable
signing key; they do not use production signing secrets.

The companion is uploaded only as `ci-android-instrumentation`, with seven-day
retention. It is outside `release-*` publication inputs and the public package
checksums. Windows/Linux packaging and application behavior are unchanged.
Publication gates (`build_only`, `confirm_publish`, repository authorization
and the protected publishing environment) remain unchanged.

After PR CI and post-merge CI pass, repeat the build-only release rehearsal on
the new main commit before preparing a beta tag. Physical drives and encrypted
Blu-ray compatibility remain separate validation requirements.
