//go:build windows || linux

package main

import (
	"strings"
	"testing"
	"time"
)

func TestRemuxBluRayPlaylistRequiresExplicitChoice(t *testing.T) {
	// Real-world BD50: the 15-hour looping decoy 00152 is longer than 00800.
	// The chooser must require a selection, not use longestTitle.
	playlists := []titleInfo{
		{Number: 152, Duration: 15*time.Hour + 54*time.Second, ChapterCount: 900},
		{Number: 612, Duration: 3*time.Hour + 22*time.Second, ChapterCount: 20},
		{Number: 800, Duration: 3*time.Hour + 22*time.Second, ChapterCount: 20},
		{Number: 1056, Duration: 15*time.Hour + 16*time.Second, ChapterCount: 900},
	}
	if _, err := selectedRemuxBluRayPlaylist(playlists, -1); err == nil {
		t.Fatal("REMUX/DEMUX accepted an unselected playlist (must not default to longest)")
	}
	for _, number := range []int{612, 800, 152} {
		got, err := selectedRemuxBluRayPlaylist(playlists, number)
		if err != nil || got.Number != number {
			t.Fatalf("explicit playlist %05d selected %v; err %v", number, got, err)
		}
	}
	if _, err := selectedRemuxBluRayPlaylist(playlists, 799); err == nil {
		t.Fatal("REMUX/DEMUX accepted nonexistent playlist 00799")
	}
	for _, number := range []int{-1, -100} {
		if _, err := selectedRemuxBluRayPlaylist(playlists, number); err == nil {
			t.Fatalf("invalid playlist %d accepted", number)
		}
	}
	if _, err := selectedRemuxBluRayPlaylist(playlists, 0); err == nil {
		t.Fatal("accepted missing playlist zero when it is not among the discovered playlists")
	}
	zero := append(playlists, titleInfo{Number: 0, Duration: time.Minute})
	if _, err := selectedRemuxBluRayPlaylist(zero, 0); err != nil {
		t.Fatalf("valid playlist 00000 should be selectable explicitly: %v", err)
	}
	label := remuxBluRayPlaylistLabel(playlists[2])
	if !strings.HasPrefix(label, "Playlist 00800 — ") || !strings.HasSuffix(label, " — 20 chapters") {
		t.Fatalf("Blu-ray movie label missing playlist number or chapters: %q", label)
	}
	decoy := remuxBluRayPlaylistLabel(playlists[0])
	if !strings.HasPrefix(decoy, "Playlist 00152 — ") || !strings.HasSuffix(decoy, " — 900 chapters") {
		t.Fatalf("fake-long playlist label missing chapter count: %q", decoy)
	}
	if label == decoy {
		t.Fatal("different Blu-ray playlist chapter counts should be visible in the chooser")
	}
	zeroLabel := remuxBluRayPlaylistLabel(titleInfo{Number: 0, Duration: time.Minute, ChapterCount: 1})
	if !strings.HasPrefix(zeroLabel, "Playlist 00000 — ") || !strings.HasSuffix(zeroLabel, " — 1 chapters") {
		t.Fatalf("valid playlist 00000 label missing chapter count: %q", zeroLabel)
	}
}

func TestRemuxBluRayPlaylistCancelCannotSelectAnything(t *testing.T) {
	// The UI maps an empty chooser selection to 0; it must be rejected.
	titles := []titleInfo{{Number:152,Duration:15*time.Hour},{Number:800,Duration:3*time.Hour}}
	_, err := selectedRemuxBluRayPlaylist(titles, -1)
	if err == nil || !strings.Contains(err.Error(), "choose") {
		t.Fatalf("cancel / empty selection should be rejected: %v", err)
	}
	// An explicit choice is accepted regardless of whether it is the longest.
	chosen, err := selectedRemuxBluRayPlaylist(titles, 800)
	if err != nil || chosen.Number != 800 {
		t.Fatalf("explicit movie playlist was not retained: %v %v", chosen, err)
	}
}
