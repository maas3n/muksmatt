//go:build windows || linux

package main

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestAdvancedBlurayArgsMixedWithOtherMedia(t *testing.T) {
	root := t.TempDir()
	if err := mkdirBlurayTestRoot(root); err != nil {
		t.Fatal(err)
	}
	rows := []mergerStream{
		{Path: "ordinary.mkv", Track: trackOption{Index: 3, Kind: "audio", Codec: "ac3"}},
		{Path: root, Track: trackOption{Index: 0, Kind: "video", Codec: "h264"}, BluRay: true, BluRayPlaylist: 800},
		{Path: root, Track: trackOption{Index: 1, Kind: "audio", Codec: "pcm_bluray"}, BluRay: true, BluRayPlaylist: 800},
		{Path: root, Track: trackOption{Index: 2, Kind: "audio", Codec: "dts"}, BluRay: true, BluRayPlaylist: 800},
	}
	args, err := mergerArgs(rows, "", "merged.mkv")
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "-playlist 800 -i bluray:"+filepath.ToSlash(root)) {
		t.Fatalf("missing direct Blu-ray input: %s", joined)
	}
	if !strings.Contains(joined, "-c copy -c:a:1 flac -f matroska merged.mkv") {
		t.Fatalf("expected PCM_BLURAY at output audio #1 to use lossless FLAC: %s", joined)
	}
	if strings.Contains(joined, "100M") || strings.Contains(joined, "-f dvdvideo") {
		t.Fatalf("DVD-only input flags leaked into Blu-ray/MKV mix: %s", joined)
	}
	if !reflect.DeepEqual(mergerBlurayLPCMAudioIndexes(rows), []int{1}) {
		t.Fatalf("wrong Blu-ray LPCM audio mapping")
	}
}

func TestAdvancedBlurayRejectsConflictingPlaylists(t *testing.T) {
	rows := []mergerStream{
		{Path: "/movies/BDMV", Track: trackOption{Index: 0, Kind: "video"}, BluRay: true, BluRayPlaylist: 0},
		{Path: "/movies/BDMV", Track: trackOption{Index: 1, Kind: "audio"}, BluRay: true, BluRayPlaylist: 800},
	}
	if _, err := mergerBluraySelections(rows); err == nil {
		t.Fatal("accepted mixed playlists for one Blu-ray")
	}
}

func TestBlurayGUIOutputNaming(t *testing.T) {
	src := bluraySource{BaseName: "My Blu-ray"}
	path := guiBlurayOutputPath(src, 0, "/tmp/output")
	if filepath.Base(path) != "My Blu-ray-bluray-00000.mkv" {
		t.Fatalf("unexpected path %q", path)
	}
}

func mkdirBlurayTestRoot(root string) error {
	return os.MkdirAll(filepath.Join(root, "BDMV", "PLAYLIST"), 0755)
}
