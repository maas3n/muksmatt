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
# Keep stable DVD libudfread 1.1.2 untouched. libbluray 1.5.0 requires
# libudfread >= 1.2.0, so build a private SONAME for Blu-ray only.
UDF_VERSION=1.2.0
UDF_SHA256=bb477cbd4cfbfc7787d9d05b71ee5e70430f5cfebf1297497f7e83547958050f
UDF_ARCHIVE="$WORK/libudfread-$UDF_VERSION.tar.xz"
UDF_SOURCE="$WORK/libudfread-$UDF_VERSION"
curl --fail --location --retry 3 --output "$UDF_ARCHIVE" \
  "https://download.videolan.org/pub/videolan/libudfread/libudfread-$UDF_VERSION.tar.xz"
printf '%s  %s\n' "$UDF_SHA256" "$UDF_ARCHIVE" | sha256sum -c --strict
tar -xf "$UDF_ARCHIVE" -C "$WORK"
cp "$UDF_SOURCE/COPYING" "$ASSETS/LIBBLURAY_UDFREAD_COPYING.txt"

build_abi() {
    local abi="$1" target="$2"
    local cc="$TOOLCHAIN/bin/${target}${API}-clang"
    local prefix="$WORK/install-${abi}"
    local build="$WORK/build-libbluray-${abi}"
    local jni="$JNI_ROOT/$abi"
    mkdir -p "$build" "$jni"
    test -f "$prefix/lib/libudfread.so"
    # VideoLAN libbluray 1.5.0 uses Meson rather than Autotools.
    # Restrict pkg-config discovery to the already cross-built Android UDF lib.
    local family
    case "$abi" in
      arm64-v8a) family="aarch64" ;;
      x86_64) family="x86_64" ;;
      *) echo "Unsupported Blu-ray ABI: $abi" >&2; exit 1 ;;
    esac
    local cross="$WORK/libbluray-$abi-cross.ini"
    cat > "$cross" <<CROSS
[binaries]
c = '$cc'
ar = '$TOOLCHAIN/bin/llvm-ar'
strip = '$TOOLCHAIN/bin/llvm-strip'
pkg-config = '/usr/bin/pkg-config'

[host_machine]
system = 'android'
cpu_family = '$family'
cpu = '$family'
endian = 'little'

[properties]
needs_exe_wrapper = true

[built-in options]
c_args = ['-O2', '-fPIC', '-I$prefix/include']
c_link_args = ['-Wl,-z,max-page-size=16384']
CROSS
    # Build the separate 1.2.0 UDF library with the same NDK cross file.
    local udf_prefix="$WORK/bluray-udf-$abi"
    local udf_build="$WORK/build-bluray-udf-$abi"
    meson setup "$udf_build" "$UDF_SOURCE" --cross-file "$cross" \
      --prefix "$udf_prefix" --libdir lib --default-library shared
    meson compile -C "$udf_build" -j2
    meson install -C "$udf_build"
    local udf_lib="$udf_prefix/lib/libudfread.so"
    test -f "$udf_lib"
    cp "$(readlink -f "$udf_lib")" "$jni/libmuksmatt_bluray_udfread.so"
    patchelf --set-soname libmuksmatt_bluray_udfread.so "$jni/libmuksmatt_bluray_udfread.so"

    PKG_CONFIG_LIBDIR="$udf_prefix/lib/pkgconfig" \
      meson setup "$build" "$SOURCE" --cross-file "$cross" \
        --prefix "$prefix" --libdir lib --default-library shared \
        -Dembed_udfread=false -Dfreetype=disabled -Dfontconfig=disabled \
        -Dlibxml2=disabled -Dbdj_jar=disabled -Denable_tools=false
    meson compile -C "$build" -j2
    meson install -C "$build"
    local lib="$prefix/lib/libbluray.so"
    test -f "$lib"
    cp "$(readlink -f "$lib")" "$jni/libbluray.so"
    patchelf --set-soname libbluray.so "$jni/libbluray.so"
    while IFS= read -r dep; do
        case "$dep" in
            libudfread.so.*) patchelf --replace-needed "$dep" libmuksmatt_bluray_udfread.so "$jni/libbluray.so" ;;
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
        "$SCRIPT_DIR/bluray_saf_files.c" \
        -L"$jni" -L"$udf_prefix/lib" -Wl,--no-as-needed \
        -lbluray -ludfread -llog -ldl \
        -Wl,-z,max-page-size=16384 -Wl,--no-undefined \
        -Wl,-soname,libmuksmatt_bluray.so \
        -o "$jni/libmuksmatt_bluray.so"
    while IFS= read -r dep; do
        case "$dep" in
            libudfread.so.*) patchelf --replace-needed "$dep" libmuksmatt_bluray_udfread.so "$jni/libmuksmatt_bluray.so" ;;
            libbluray.so.*) patchelf --replace-needed "$dep" libbluray.so "$jni/libmuksmatt_bluray.so" ;;
        esac
    done < <(patchelf --print-needed "$jni/libmuksmatt_bluray.so")
    patchelf --print-needed "$jni/libmuksmatt_bluray.so" | grep -q '^libbluray.so$'
    "$TOOLCHAIN/bin/llvm-nm" -D "$jni/libmuksmatt_bluray.so" |
        grep -q 'Java_io_github_maas3n_mattmux_BlurayNativeIsoBridge_nativeInspectIso'
    "$TOOLCHAIN/bin/llvm-nm" -D "$jni/libmuksmatt_bluray.so" |
        grep -q 'Java_io_github_maas3n_mattmux_BluraySafTreeBridge_nativeInspectTree'
    "$TOOLCHAIN/bin/llvm-strip" --strip-unneeded "$jni/libbluray.so" "$jni/libmuksmatt_bluray.so" "$jni/libmuksmatt_bluray_udfread.so"
    {
      echo
      echo "Blu-ray Android native ISO reader ($abi): libbluray $BLURAY_VERSION"
      echo "Blu-ray source SHA256: $BLURAY_SHA256"
      echo "Playlist scanning and bounded transport sample via seekable SAF descriptor"
      for file in "$jni/libbluray.so" "$jni/libmuksmatt_bluray.so" "$jni/libmuksmatt_bluray_udfread.so"; do
          echo "$(basename "$file"):"
          patchelf --print-needed "$file" | sed 's/^/  needs: /'
          sha256sum "$file" | sed 's/^/  sha256: /'
      done
    } >> "$ASSETS/ffmpeg-build-info.txt"
}

build_abi arm64-v8a aarch64-linux-android
build_abi x86_64 x86_64-linux-android

mkdir -p "$ROOT/dist/android-release"
cp "$ARCHIVE" "$UDF_ARCHIVE" "$ROOT/dist/android-release/"
cp "$ASSETS/LIBBLURAY_COPYING.txt" "$ASSETS/LIBBLURAY_UDFREAD_COPYING.txt" "$ROOT/dist/android-release/"
echo "Android native libbluray ISO reader built for both ABIs."
