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
		{Number: 152, Duration: 15*time.Hour + 54*time.Second},
		{Number: 612, Duration: 3*time.Hour + 22*time.Second},
		{Number: 800, Duration: 3*time.Hour + 22*time.Second},
		{Number: 1056, Duration: 15*time.Hour + 16*time.Second},
	}
	if _, err := selectedRemuxBluRayPlaylist(playlists, 0); err == nil {
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
	label := remuxBluRayPlaylistLabel(800, 3*time.Hour+22*time.Second)
	if !strings.HasPrefix(label, "Playlist 00800 — ") {
		t.Fatalf("playlist label lost five-digit ID: %q", label)
	}
}

func TestRemuxBluRayPlaylistCancelCannotSelectAnything(t *testing.T) {
	// The UI maps an empty chooser selection to 0; it must be rejected.
	titles := []titleInfo{{Number:152,Duration:15*time.Hour},{Number:800,Duration:3*time.Hour}}
	_, err := selectedRemuxBluRayPlaylist(titles, 0)
	if err == nil || !strings.Contains(err.Error(), "choose") {
		t.Fatalf("cancel / empty selection should be rejected: %v", err)
	}
	// An explicit choice is accepted regardless of whether it is the longest.
	chosen, err := selectedRemuxBluRayPlaylist(titles, 800)
	if err != nil || chosen.Number != 800 {
		t.Fatalf("explicit movie playlist was not retained: %v %v", chosen, err)
	}
}
