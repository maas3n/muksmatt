#!/usr/bin/env bash
# End-to-end desktop Blu-ray CLI test on self-authored, unencrypted BDMV/ISO.
# Uses the public cliBluray entry point, real libbluray, FFmpeg and FFprobe.
# No commercial media, AACS/BD+ credentials or signing keys are involved.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CLI="${1:?Usage: bash scripts/test-bluray-desktop-cli.sh /path/to/muksmatt-cli /path/to/muksmatt-bluray-nav}"
NAV="${2:?Expected compiled muksmatt-bluray-nav path}"
test -x "$CLI" && test -x "$NAV"
for bin in ffmpeg ffprobe python3; do command -v "$bin" >/dev/null; done
export PATH="$(dirname "$NAV"):$PATH"
test "$("$NAV" --version)" = MUKSMATT_BD_NAV_1
for bin in ffmpeg ffprobe; do
    "$bin" -hide_banner -protocols 2>&1 | grep -Eq '^[[:space:]]*bluray$' || {
        echo "$bin was built without the Blu-ray protocol" >&2
        exit 1
    }
done

WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT
DISC="$WORK/authored-disc"
bash "$ROOT/android/native/tests/test-bluray-media.sh" --fixtures-only "$DISC"
ISO="$DISC/movie.iso"
test -s "$ISO" && test -d "$DISC/BDMV/PLAYLIST"

"$CLI" bluray scan "$DISC" > "$WORK/folder-scan.txt"
"$CLI" bluray scan "$ISO" > "$WORK/iso-scan.txt"
grep -F longest "$WORK/folder-scan.txt"
grep -F longest "$WORK/iso-scan.txt"

# Parse the selected playlist from the public CLI output, not an assumed ID.
PLAYLIST="$(awk '$NF == "longest" { print $1; exit }' "$WORK/folder-scan.txt")"
[[ "$PLAYLIST" =~ ^[0-9]{5}$ ]] || { echo "No longest Blu-ray playlist in scan" >&2; exit 1; }
PLAYLIST_NUMBER="$((10#$PLAYLIST))"

# Independently discover the absolute FFmpeg source indexes for -streams.
python3 - "$DISC" "$PLAYLIST_NUMBER" "$WORK/selected-indices.txt" <<'PY'
import json, subprocess, sys
src, playlist, destination = sys.argv[1:]
info = json.loads(subprocess.check_output([
    "ffprobe", "-v", "error", "-playlist", playlist, "-show_streams",
    "-of", "json", "bluray:" + src], text=True))
video = [s["index"] for s in info["streams"] if s.get("codec_type") == "video"]
audio = [s["index"] for s in info["streams"] if s.get("codec_type") == "audio"]
assert len(video) == len(audio) == 1, info
with open(destination, "w", encoding="utf-8") as fp:
    fp.write(f"{video[0]} {audio[0]}\n")
PY
read -r VIDEO_INDEX AUDIO_INDEX < "$WORK/selected-indices.txt"

"$CLI" bluray remux --output "$WORK/folder.mkv" "$DISC"
"$CLI" bluray remux --output "$WORK/iso.mkv" "$ISO"
"$CLI" bluray remux --playlist "$PLAYLIST" --streams "$VIDEO_INDEX" \
    --no-chapters --output "$WORK/video-only.mkv" "$DISC"
"$CLI" bluray remux --playlist "$PLAYLIST" --streams "$AUDIO_INDEX" \
    --no-chapters --output "$WORK/audio-only.mkv" "$ISO"

# Validate Matroska tracks, mandatory LPCM->FLAC and chapter policy.
python3 - "$WORK" <<'PY'
import json, pathlib, subprocess, sys
work = pathlib.Path(sys.argv[1])
def probe(name):
    data = subprocess.check_output([
        "ffprobe", "-v", "error", "-show_streams", "-show_chapters",
        "-of", "json", str(work / name)], text=True)
    return json.loads(data)
for name in ("folder.mkv", "iso.mkv"):
    data = probe(name)
    kinds = [(s["codec_type"], s["codec_name"]) for s in data["streams"]]
    assert kinds == [("video", "mpeg2video"), ("audio", "flac")], (name, kinds)
    audio = data["streams"][1]
    assert int(audio["sample_rate"]) == 48000 and audio["channels"] == 2, audio
    assert int(audio.get("bits_per_raw_sample", 24)) == 24, audio
    assert len(data.get("chapters", [])) >= 2, (name, data.get("chapters"))
v = probe("video-only.mkv")
assert [(s["codec_type"], s["codec_name"]) for s in v["streams"]] == [
    ("video", "mpeg2video")], v
assert not v.get("chapters"), v["chapters"]
a = probe("audio-only.mkv")
assert [(s["codec_type"], s["codec_name"]) for s in a["streams"]] == [
    ("audio", "flac")], a
assert not a.get("chapters"), a["chapters"]
PY

# Decode the source playlist and both outputs, checking every frame and PCM
# sample. This catches accidental transcoding, dropped streams and bit-depth
# loss that stream-count checks alone could miss.
sha() { sha256sum | awk '{print $1}'; }
source_video="$(ffmpeg -hide_banner -loglevel error -nostdin -playlist "$PLAYLIST_NUMBER" \
    -i "bluray:$DISC" -map 0:v:0 -pix_fmt yuv420p -f rawvideo - | sha)"
source_audio="$(ffmpeg -hide_banner -loglevel error -nostdin -playlist "$PLAYLIST_NUMBER" \
    -i "bluray:$DISC" -map 0:a:0 -c:a pcm_s24le -f s24le - | sha)"
test -n "$source_video" && test -n "$source_audio"
for name in folder iso video-only; do
    got="$(ffmpeg -hide_banner -loglevel error -nostdin -i "$WORK/$name.mkv" \
      -map 0:v:0 -pix_fmt yuv420p -f rawvideo - | sha)"
    test "$source_video" = "$got" || { echo "Video parity failed: $name" >&2; exit 1; }
done
for name in folder iso audio-only; do
    got="$(ffmpeg -hide_banner -loglevel error -nostdin -i "$WORK/$name.mkv" \
      -map 0:a:0 -c:a pcm_s24le -f s24le - | sha)"
    test "$source_audio" = "$got" || { echo "LPCM/FLAC sample parity failed: $name" >&2; exit 1; }
done

# Failed explicit selection must leave no visible final MKV.
if "$CLI" bluray remux --streams 999 --output "$WORK/invalid.mkv" "$DISC" > "$WORK/invalid.log" 2>&1; then
    echo "Nonexistent Blu-ray stream index was accepted" >&2
    exit 1
fi
test ! -e "$WORK/invalid.mkv"
echo "Desktop Blu-ray BDMV/ISO CLI scan, playlist, stream-selection, chapters, frame parity and lossless FLAC PASSED"
