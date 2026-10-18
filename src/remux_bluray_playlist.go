//go:build windows || linux

package main

import (
	"fmt"
	"time"
)

// An explicit playlist number from the user is required for REMUX/DEMUX
// Blu-ray scan. Never fall back to "longest" on an empty choice, since fake
// looping playlists can be many hours longer than the actual film.
func selectedRemuxBluRayPlaylist(titles []titleInfo, requested int) (titleInfo, error) {
	if requested < 1 {
		return titleInfo{}, fmt.Errorf("choose a Blu-ray playlist before selecting streams")
	}
	for _, title := range titles {
		if title.Number == requested {
			return title, nil
		}
	}
	return titleInfo{}, fmt.Errorf("Blu-ray playlist %05d is no longer available; scan again", requested)
}

func remuxBluRayPlaylistLabel(number int, duration time.Duration) string {
	return fmt.Sprintf("Playlist %05d — %s", number, formatDuration(duration))
}
