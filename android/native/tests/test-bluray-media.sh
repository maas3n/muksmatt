#!/usr/bin/env bash
# Unencrypted self-authored Blu-ray BDMV+ISO validation for same Android C core.
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT
VERSION=2.7.0
curl -fLsS --retry 3 --proto '=https' --tlsv1.2 \
  "https://github.com/justdan96/tsMuxer/releases/download/$VERSION/tsMuxer-$VERSION-linux.zip" \
  -o "$WORK/tsmuxer.zip"
echo "Fixture tool SHA-256 (pin after review):"
sha256sum "$WORK/tsmuxer.zip"
mkdir "$WORK/tools"
unzip -q "$WORK/tsmuxer.zip" -d "$WORK/tools"
TSMUXER="$(find "$WORK/tools" -type f \( -iname 'tsmuxer' -o -iname 'tsmuxer_linux' \) -print -quit)"
test -n "$TSMUXER"
chmod +x "$TSMUXER"

ffmpeg -hide_banner -loglevel error -nostdin -y \
  -f lavfi -i "testsrc2=size=1280x720:rate=25:duration=6" \
  -an -c:v mpeg2video -pix_fmt yuv420p -b:v 4000k -g 12 "$WORK/video.m2v"
ffmpeg -hide_banner -loglevel error -nostdin -y \
  -f lavfi -i "sine=frequency=733:sample_rate=48000:duration=6" \
  -ac 2 -c:a pcm_s24le "$WORK/audio.wav"
cat > "$WORK/disc.meta" <<META
MUXOPT --blu-ray --custom-chapters=00:00:01.000;00:00:03.000
V_MPEG-2, $WORK/video.m2v, fps=25
A_LPCM, $WORK/audio.wav, lang=eng
META
"$TSMUXER" "$WORK/disc.meta" "$WORK/disc"
test -d "$WORK/disc/BDMV/PLAYLIST"
test -d "$WORK/disc/BDMV/STREAM"
"$TSMUXER" "$WORK/disc.meta" "$WORK/disc.iso"
test -s "$WORK/disc.iso"

cc -std=c11 -O2 -Wall -Wextra \
  "$ROOT/android/native/bluray_mkv_core.c" \
  "$ROOT/android/native/tests/bluray_mkv_media_harness.c" \
  -o "$WORK/bluray-remux-test" \
  $(pkg-config --cflags --libs libbluray libavformat libavcodec libavutil libswresample)

"$WORK/bluray-remux-test" "$WORK/disc" "$WORK/folder.mkv" all
"$WORK/bluray-remux-test" "$WORK/disc.iso" "$WORK/image.mkv" all
"$WORK/bluray-remux-test" "$WORK/disc" "$WORK/audio-only.mkv" audio
"$WORK/bluray-remux-test" "$WORK/disc" "$WORK/video-only.mkv" video

python3 - "$WORK" <<'PY'
import json, pathlib, subprocess, sys
work = pathlib.Path(sys.argv[1])
def probe(path):
    p = subprocess.run(["ffprobe","-v","error","-show_streams","-show_chapters",
                        "-of","json",str(path)], check=True, capture_output=True,text=True)
    return json.loads(p.stdout)
for name in ("folder.mkv", "image.mkv"):
    data = probe(work/name)
    streams=data["streams"]
    codecs={(s["codec_type"],s["codec_name"]) for s in streams}
    print(name,codecs,"chapters",len(data.get("chapters",[])))
    assert ("video","mpeg2video") in codecs, codecs
    assert ("audio","flac") in codecs, codecs
    assert len(data.get("chapters",[])) >= 2, data.get("chapters")
    for s in streams:
        if s["codec_type"]=="audio":
            assert s["sample_rate"]=="48000" and s["channels"]==2, s
audio=probe(work/"audio-only.mkv")["streams"]
video=probe(work/"video-only.mkv")["streams"]
assert any(s["codec_name"]=="flac" for s in audio)
assert not any(s["codec_type"]=="video" for s in audio)
assert any(s["codec_type"]=="video" for s in video)
assert not any(s["codec_type"]=="audio" for s in video)
PY

for name in folder image audio-only; do
  ffmpeg -hide_banner -loglevel error -nostdin -y -i "$WORK/$name.mkv" \
    -map 0:a:0 -f s24le -c:a pcm_s24le "$WORK/$name.pcm"
done
ffmpeg -hide_banner -loglevel error -nostdin -y -i "$WORK/audio.wav" \
  -map 0:a:0 -f s24le -c:a pcm_s24le "$WORK/reference.pcm"
python3 - "$WORK" <<'PY'
import hashlib,pathlib,sys
d=pathlib.Path(sys.argv[1])
ref=(d/"reference.pcm").read_bytes()
print("Reference decoded PCM:",len(ref),hashlib.sha256(ref).hexdigest())
for name in ("folder.pcm","image.pcm","audio-only.pcm"):
    pcm=(d/name).read_bytes()
    print(name,len(pcm),hashlib.sha256(pcm).hexdigest())
    if pcm != ref:
        raise SystemExit("Lossless LPCM -> FLAC sample parity failure: "+name)
PY
echo "Blu-ray ISO/BDMV native media validation PASSED"
