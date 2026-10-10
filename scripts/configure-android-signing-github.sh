#!/usr/bin/env bash
# Upload the existing muKsMaTT signing identities without copying secrets to chat or Git.
set +x
set -euo pipefail
umask 077

repo="maas3n/muksmatt"
keys="$HOME/muksmatt-android-signing"

for tool in gh openssl sha256sum; do
  command -v "$tool" >/dev/null || { echo "Missing required command: $tool" >&2; exit 1; }
done
for file in muksmatt-app-signing.jks.base64 muksmatt-app-signing-cert.pem muksmatt-upload.jks.base64 muksmatt-upload-cert.pem; do
  test -s "$keys/$file" || { echo "Missing signing file: $keys/$file" >&2; exit 1; }
done
gh auth status -h github.com >/dev/null

# Pin to the identities generated on 2026-10-10; abort on a mismatch.
app_sha="$(openssl x509 -in "$keys/muksmatt-app-signing-cert.pem" -outform DER | sha256sum | awk '{print $1}')"
upload_sha="$(openssl x509 -in "$keys/muksmatt-upload-cert.pem" -outform DER | sha256sum | awk '{print $1}')"
test "$app_sha" = "752de89191e74fc7c4336403c7ad07006f0a73980e3fc48f0d89d010e68a09f5" || {
  echo "App-signing certificate does not match the intended identity." >&2; exit 1;
}
test "$upload_sha" = "2c093afc79fc2c4290985637d8e7730746c38a032f4d368aed5e80b86b64744f" || {
  echo "Play upload certificate does not match the intended identity." >&2; exit 1;
}
test "$app_sha" != "$upload_sha" || { echo "Signing identities must be distinct." >&2; exit 1; }

printf '\nThe private keystores and passwords stay on this computer; GitHub CLI uploads them directly to GitHub Actions secrets.\n'
read -r -s -p "App-signing keystore password: " app_store; echo
read -r -s -p "App-signing private-key password: " app_key; echo
read -r -s -p "Play upload keystore password: " upload_store; echo
read -r -s -p "Play upload private-key password: " upload_key; echo
trap 'unset app_store app_key upload_store upload_key' EXIT
for password in "$app_store" "$app_key" "$upload_store" "$upload_key"; do
  test "$(printf '%s' "$password" | wc -c)" -ge 12 || { echo "Password cannot be blank or shorter than 12 characters." >&2; exit 1; }
done

# Preserve any existing environment protection rules; create only if absent.
for env in android-release android-play; do
  if ! gh api "repos/$repo/environments/$env" >/dev/null 2>&1; then
    gh api --method PUT "repos/$repo/environments/$env" >/dev/null
  fi
done

gh secret set ANDROID_APP_SIGNING_KEYSTORE_BASE64 --repo "$repo" --env android-release < "$keys/muksmatt-app-signing.jks.base64"
printf '%s' "$app_store" | gh secret set ANDROID_APP_SIGNING_STORE_PASSWORD --repo "$repo" --env android-release
printf '%s' muksmatt-app-signing | gh secret set ANDROID_APP_SIGNING_KEY_ALIAS --repo "$repo" --env android-release
printf '%s' "$app_key" | gh secret set ANDROID_APP_SIGNING_KEY_PASSWORD --repo "$repo" --env android-release

gh secret set ANDROID_UPLOAD_KEYSTORE_BASE64 --repo "$repo" --env android-play < "$keys/muksmatt-upload.jks.base64"
printf '%s' "$upload_store" | gh secret set ANDROID_UPLOAD_STORE_PASSWORD --repo "$repo" --env android-play
printf '%s' muksmatt-upload | gh secret set ANDROID_UPLOAD_KEY_ALIAS --repo "$repo" --env android-play
printf '%s' "$upload_key" | gh secret set ANDROID_UPLOAD_KEY_PASSWORD --repo "$repo" --env android-play

gh variable set ANDROID_APP_SIGNING_CERT_SHA256 --repo "$repo" --body "$app_sha"
gh variable set ANDROID_UPLOAD_CERT_SHA256 --repo "$repo" --body "$upload_sha"

# Intentionally do not set MUKSMATT_RELEASE_ENABLED or run a release workflow.
echo
echo "Signing secrets and public certificate fingerprints uploaded; production publishing remains locked."
echo "Dispatching the signing-only self-test (no APK or public release):"
gh workflow run android-signing-check.yml --repo "$repo" --ref main
echo "Check progress at https://github.com/maas3n/muksmatt/actions/workflows/android-signing-check.yml"
