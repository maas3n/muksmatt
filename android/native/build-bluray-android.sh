#!/usr/bin/env bash
# Phase 5: Android libbluray ISO reader for both ABIs; no DVD engine changes.
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
WORK="${SCRIPT_DIR}/.work"
ASSETS="${ROOT}/android/app/src/main/assets/ffmpeg"
JNI_ROOT="${ROOT}/android/app/src/main/jniLibs"

: "${ANDROID_NDK_HOME:?ANDROID_NDK_HOME is required}"
BLURAY_VERSION=1.5.0
BLURAY_SHA256=f676408e91a5d321abf8b8d4dfdae36205c297dab5c54c3ec519639025f474a2
BLURAY_URL="https://download.videolan.org/pub/videolan/libbluray/${BLURAY_VERSION}/libbluray-${BLURAY_VERSION}.tar.xz"
ARCHIVE="${WORK}/libbluray-${BLURAY_VERSION}.tar.xz"
SOURCE="${WORK}/libbluray-${BLURAY_VERSION}"
TOOLCHAIN="${ANDROID_NDK_HOME}/toolchains/llvm/prebuilt/linux-x86_64"
API="${ANDROID_API:-26}"
test -d "$TOOLCHAIN"
mkdir -p "$WORK" "$ASSETS"
curl --fail --location --retry 3 --output "$ARCHIVE" "$BLURAY_URL"
printf '%s  %s\n' "$BLURAY_SHA256" "$ARCHIVE" | sha256sum -c --strict
tar -xf "$ARCHIVE" -C "$WORK"
cp "$SOURCE/COPYING" "$ASSETS/LIBBLURAY_COPYING.txt"

build_abi() {
    local abi="$1" target="$2"
    local cc="$TOOLCHAIN/bin/${target}${API}-clang"
    local prefix="$WORK/install-${abi}"
    local build="$WORK/build-libbluray-${abi}"
    local jni="$JNI_ROOT/$abi"
    mkdir -p "$build" "$jni"
    test -f "$prefix/lib/libudfread.so"
    (
      cd "$build"
      CC="$cc" AR="$TOOLCHAIN/bin/llvm-ar" \
        RANLIB="$TOOLCHAIN/bin/llvm-ranlib" \
        STRIP="$TOOLCHAIN/bin/llvm-strip" \
        PKG_CONFIG_PATH="$prefix/lib/pkgconfig" \
        CPPFLAGS="-I$prefix/include" \
        CFLAGS="-O2 -fPIC" \
        LDFLAGS="-L$prefix/lib -Wl,-z,max-page-size=16384" \
        "$SOURCE/configure" --host="$target" --prefix="$prefix" \
          --disable-bdjava --without-freetype --without-libxml2 \
          --enable-shared --disable-static
      make -j2
      make install
    )
    local lib="$prefix/lib/libbluray.so"
    test -f "$lib"
    cp "$(readlink -f "$lib")" "$jni/libbluray.so"
    patchelf --set-soname libbluray.so "$jni/libbluray.so"
    while IFS= read -r dep; do
        case "$dep" in
            libudfread.so.*) patchelf --replace-needed "$dep" libudfread.so "$jni/libbluray.so" ;;
        esac
    done < <(patchelf --print-needed "$jni/libbluray.so")
    local needed
    needed="$(patchelf --print-needed "$jni/libbluray.so")"
    if grep -Eiq '(libjvm|libxml2|libfreetype|libdvdnav|libdvdread)' <<< "$needed"; then
        echo "Unexpected Android Blu-ray native dependency: $needed" >&2
        exit 1
    fi

    "$cc" -O2 -fPIC -shared \
        -I"$prefix/include" \
        "$SCRIPT_DIR/bluray_saf_blocks.c" \
        "$SCRIPT_DIR/bluray_iso_jni.c" \
        -L"$jni" -L"$prefix/lib" -Wl,--no-as-needed \
        -lbluray -ludfread -llog -ldl \
        -Wl,-z,max-page-size=16384 -Wl,--no-undefined \
        -Wl,-soname,libmuksmatt_bluray.so \
        -o "$jni/libmuksmatt_bluray.so"
    while IFS= read -r dep; do
        case "$dep" in
            libudfread.so.*) patchelf --replace-needed "$dep" libudfread.so "$jni/libmuksmatt_bluray.so" ;;
            libbluray.so.*) patchelf --replace-needed "$dep" libbluray.so "$jni/libmuksmatt_bluray.so" ;;
        esac
    done < <(patchelf --print-needed "$jni/libmuksmatt_bluray.so")
    patchelf --print-needed "$jni/libmuksmatt_bluray.so" | grep -q '^libbluray.so$'
    "$TOOLCHAIN/bin/llvm-nm" -D "$jni/libmuksmatt_bluray.so" |
        grep -q 'Java_io_github_maas3n_mattmux_BlurayNativeIsoBridge_nativeInspectIso'
    "$TOOLCHAIN/bin/llvm-strip" --strip-unneeded "$jni/libbluray.so" "$jni/libmuksmatt_bluray.so"
    {
      echo
      echo "Blu-ray Android native ISO reader ($abi): libbluray $BLURAY_VERSION"
      echo "Blu-ray source SHA256: $BLURAY_SHA256"
      echo "Playlist scanning and bounded transport sample via seekable SAF descriptor"
      for file in "$jni/libbluray.so" "$jni/libmuksmatt_bluray.so"; do
          echo "$(basename "$file"):"
          patchelf --print-needed "$file" | sed 's/^/  needs: /'
          sha256sum "$file" | sed 's/^/  sha256: /'
      done
    } >> "$ASSETS/ffmpeg-build-info.txt"
}

build_abi arm64-v8a aarch64-linux-android
build_abi x86_64 x86_64-linux-android

mkdir -p "$ROOT/dist/android-release"
cp "$ARCHIVE" "$ROOT/dist/android-release/"
cp "$ASSETS/LIBBLURAY_COPYING.txt" "$ROOT/dist/android-release/"
echo "Android native libbluray ISO reader built for both ABIs."
