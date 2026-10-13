# muKsMaTT — Google Play preparation (not published)

**muKsMaTT v0.1.0-beta.1 is a public GitHub prerelease APK, not a Google Play release.** Play Store listing, purchase verification and production rollout still require separate decisions and validation. Do not confuse upstream MattRip settings with muKsMaTT.

## Android application identity

- App name: **muKsMaTT**
- Application ID: **`io.github.maas3n.muksmatt`** (verified in `android/app/build.gradle.kts`)
- Min SDK: Android 8.0 / API 26
- Target SDK: API 36
- Supported native ABIs: `arm64-v8a` and `x86_64`
- ChromeOS: requires the Android application environment

The application ID is distinct from MattRip and MattMux. Confirm it before creating a production Play listing; after launch the installed-app identity is difficult to change.

## Current billing status

The code currently includes a Google Play Billing dependency, but `BuildConfig.ENABLE_BILLING_PURCHASES = false`. **Purchases are deliberately disabled in the public beta.**

Some inherited MattRip implementation/documentation uses the historical product identifier `mattrip_pro`. **Do not register, advertise, or charge users for that inherited SKU as though it were an authorized muKsMaTT product.** Choose and verify a distinct muKsMaTT product strategy, update and test the application consistently, and implement server-side purchase-token verification before enabling any monetization. Any pricing and rollout require separate approval.

## App-signing and upload keys

muKsMaTT maintains **two distinct signing identities** documented in [SIGNING.md](SIGNING.md):

- The **app-signing key** signs the public GitHub APK. Google Play App Signing must use the corresponding public app-signing certificate if Play-delivered APKs are to be update-compatible with direct GitHub APKs.
- The **upload key** signs Play AAB uploads only; it is not the final installed-app signing identity.

For future enrollment, follow the current Play Console key import process. Confirm that the Play **app signing certificate SHA-256** equals `ANDROID_APP_SIGNING_CERT_SHA256` and that the **upload certificate SHA-256** equals `ANDROID_UPLOAD_CERT_SHA256`. A mismatch is a release blocker. Never upload keys or credentials to GitHub issues or documentation.

## Future Play validation

1. Confirm application ID, signing identities, supported devices, privacy documentation, legal/distribution compliance and the current Google Play requirements.
2. Generate the Play AAB only through the authorized, separately signed **Android / ChromeOS Play bundle** workflow.
3. Test app startup, DVD and experimental Blu-ray behavior on actual Android phones and Chromebooks. CSS is intentionally **disabled** in the Play variant.
4. If billing is ever authorized, design a new product configuration, server-side verification, and tests for pending, failed, restored and acknowledged purchases.
5. Use internal and closed testing before any public rollout. This GitHub beta does not authorize charging customers or publishing to Play.

The relevant release-policy overview is in [RELEASING.md](../RELEASING.md).
