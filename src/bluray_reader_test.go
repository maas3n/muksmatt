//go:build windows || linux

package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func testBlurayDiscRoot(t *testing.T) bluraySource {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "Blu ray Movie")
	if err := os.MkdirAll(filepath.Join(dir, "BDMV", "PLAYLIST"), 0755); err != nil {
		t.Fatal(err)
	}
	s, err := resolveBluraySource(dir)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestLibblurayProtocolDetection(t *testing.T) {
	if !hasBlurayProtocol([]byte("Input:\n file\n bluray\nOutput:\n file\n")) {
		t.Fatal("failed to detect bluray protocol")
	}
	for _, content := range []string{"Input:\n file\n", "Input:\n notbluray\n", ""} {
		if hasBlurayProtocol([]byte(content)) {
			t.Fatalf("false positive in %q", content)
		}
	}
}

func TestBlurayInputUsesSelectedPlaylistAndNotDVDFilters(t *testing.T) {
	source := testBlurayDiscRoot(t)
	args, err := blurayInputArgs(source, blurayPlaylist{Number: 800})
	if err != nil { t.Fatal(err) }
	if len(args) != 4 || args[0] != "-playlist" || args[1] != "800" ||
		args[2] != "-i" || !strings.HasPrefix(args[3], "bluray:") ||
		!strings.Contains(args[3], "Blu ray Movie") {
		t.Fatalf("invalid bluray source flags: %v", args)
	}
	for _, forbidden := range []string{"-probesize", "-analyzeduration", "dvdvideo", "+genpts"} {
		if strings.Contains(strings.Join(args, " "), forbidden) {
			t.Fatalf("DVD-only flag leaked to Blu-ray: %s", forbidden)
		}
	}
	iso, err := resolveBluraySource(filepath.Join(t.TempDir(), "invalid.iso"))
	if err == nil || iso.Kind != 0 { t.Fatal("nonexistent ISO should fail") }
	s := bluraySource{Kind: bluraySourceISOCandidate, Input: "Disc.iso"}
	isoArgs, err := blurayInputArgs(s, blurayPlaylist{Number: 800})
	if err != nil || !strings.Contains(isoArgs[3], "Disc.iso") {
		t.Fatalf("libbluray ISO source flags: %v, err %v", isoArgs, err)
	}
}

func TestBlurayProbeStreamFilteringAndLPCMMapping(t *testing.T) {
	inventory := map[string]interface{}{"streams": []map[string]interface{}{
		{"index": 0, "codec_name": "h264", "codec_type": "video"},
		{"index": 1, "codec_name": "truehd", "codec_type": "audio"},
		{"index": 2, "codec_name": "pcm_bluray", "codec_type": "audio"},
		{"index": 3, "codec_name": "hdmv_pgs_subtitle", "codec_type": "subtitle"},
		{"index": 4, "codec_name": "bin_data", "codec_type": "data"},
		{"index": 5, "codec_name": "ac3", "codec_type": "audio"},
	}}
	data, _ := json.Marshal(inventory)
	streams, err := parseBlurayProbe(data)
	if err != nil { t.Fatal(err) }
	if len(streams) != 5 { t.Fatalf("expected 5 A/V/sub streams, got %#v", streams) }

	selected, err := selectedBlurayStreams(streams, []int{5, 0, 2})
	if err != nil { t.Fatal(err) }
	if !reflect.DeepEqual([]int{selected[0].Index, selected[1].Index, selected[2].Index}, []int{0, 2, 5}) {
		t.Fatalf("absolute stream selection did not follow probe order: %#v", selected)
	}
	for _, ids := range [][]int{{-1}, {2, 2}, {100}, {4}} {
		if _, err := selectedBlurayStreams(streams, ids); err == nil {
			t.Fatalf("accepted invalid stream indices %v", ids)
		}
	}
	source := testBlurayDiscRoot(t)
	playlist := blurayPlaylist{Number: 800}
	args, err := blurayRemuxArgs(source, playlist, streams, "chapters.ffmeta", "partial.mkv")
	if err != nil { t.Fatal(err) }
	joined := strings.Join(args, " ")
	for _, required := range []string{"-map 0:0", "-map 0:1", "-map 0:2", "-map 0:3", "-map 0:5",
		"-f ffmetadata -i chapters.ffmeta", "-map_chapters 1", "-c copy", "-c:a:1 flac", "-f matroska partial.mkv"} {
		if !strings.Contains(joined, required) {
			t.Fatalf("missing remux rule %q in %s", required, joined)
		}
	}
	if strings.Contains(joined, "-c:a:0 flac") || strings.Contains(joined, "-c:a:2 flac") {
		t.Fatal("LPCM FLAC rule applied to wrong audio stream")
	}
	if strings.Index(joined, "-c copy") > strings.LastIndex(joined, "-c:a:1 flac") {
		t.Fatal("global copy must precede per-LPCM override")
	}
	noChapters, err := blurayRemuxArgs(source, playlist, selected, "", "partial.mkv")
	if err != nil { t.Fatal(err) }
	if !strings.Contains(strings.Join(noChapters, " "), "-map_chapters -1") {
		t.Fatal("missing no-chapters rule")
	}
}

func TestBlurayChapterMetadataUses45KTimebase(t *testing.T) {
	p := blurayPlaylist{Chapters: []blurayChapter{
		{Number: 1, StartTicks: 0, EndTicks: 90000},
		{Number: 2, StartTicks: 90000, EndTicks: 135000},
	}}
	s := string(blurayFFmetadata(p))
	if !strings.HasPrefix(s, ";FFMETADATA1\n") || strings.Count(s, "[CHAPTER]") != 2 ||
		!strings.Contains(s, "TIMEBASE=1/45000\nSTART=90000\nEND=135000") {
		t.Fatalf("incorrect FFmetadata chapter timeline:\n%s", s)
	}
	if p.Chapters[1].Start() != 2*time.Second {
		t.Fatalf("unexpected chapter clock calculation: %s", p.Chapters[1].Start())
	}
}

func TestBlurayProbeRejectsInvalidInventories(t *testing.T) {
	for _, raw := range []string{
		"not json", "{\"streams\":[]}",
		"{\"streams\":[{\"index\":1,\"codec_name\":\"h264\",\"codec_type\":\"video\"},{\"index\":1,\"codec_name\":\"ac3\",\"codec_type\":\"audio\"}]}",
		"{\"streams\":[{\"index\":-1,\"codec_name\":\"h264\",\"codec_type\":\"video\"}]}",
	} {
		if _, err := parseBlurayProbe([]byte(raw)); err == nil {
			t.Fatalf("accepted invalid inventory %q", raw)
		}
	}
}
