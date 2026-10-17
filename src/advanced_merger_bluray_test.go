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

func TestAdvancedBlurayRequiresExplicitPlaylist(t *testing.T) {
	// A real BD50 exposed 15-hour looping decoys. ADVANCED must never pick
	// those just because their duration exceeds the desired movie.
	playlists := []blurayPlaylist{
		{Number: 152, DurationTicks: 45000 * 60 * 60 * 15},
		{Number: 612, DurationTicks: 45000 * (3*60*60 + 22)},
		{Number: 800, DurationTicks: 45000 * (3*60*60 + 22)},
		{Number: 1056, DurationTicks: 45000 * 60 * 60 * 15},
	}
	if _, err := advancedBlurayPlaylist(playlists, -1); err == nil {
		t.Fatal("ADVANCED silently selected longest playlist without a choice")
	}
	for _, want := range []int{612, 800} {
		got, err := advancedBlurayPlaylist(playlists, want)
		if err != nil || got.Number != want {
			t.Fatalf("playlist %05d: got %05d, err %v", want, got.Number, err)
		}
	}
	if _, err := advancedBlurayPlaylist(playlists, 799); err == nil {
		t.Fatal("ADVANCED accepted a nonexistent playlist")
	}
	if _, err := advancedBlurayPlaylist(playlists, 100000); err == nil {
		t.Fatal("ADVANCED accepted an invalid playlist number")
	}
}
