//go:build windows || linux

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

// In the M2TS variant of MPEG-TS, PGS subtitles are supported natively.
// Text subtitles such as SRT cannot be reliably stored as playable Blu-ray
// subtitle tracks. Reject rather than silently dropping the selected tracks.
func validateAdvancedM2TSStreams(streams []mergerStream) error {
	video := 0
	for _, row := range streams {
		if row.Track.Kind == "chapters" {
			continue // Explicitly excluded from M2TS outputs.
		}
		codec := strings.ToLower(row.Track.Codec)
		switch row.Track.Kind {
		case "video":
			switch codec {
			case "h264", "hevc", "mpeg2video", "vc1":
				video++
			default:
				return fmt.Errorf("M2TS cannot preserve video codec %s (%s); use MKV", row.Track.Codec, filepath.Base(row.Path))
			}
		case "audio":
			switch codec {
			case "ac3", "eac3", "dts", "truehd", "pcm_bluray":
			default:
				return fmt.Errorf("M2TS cannot preserve audio codec %s (%s); use MKV or supply a Blu-ray-compatible audio track", row.Track.Codec, filepath.Base(row.Path))
			}
		case "subtitle":
			if codec != "hdmv_pgs_subtitle" {
				return fmt.Errorf("M2TS cannot preserve %s subtitles from %s as native Blu-ray subtitles; use MKV or convert the subtitles to PGS first", row.Track.Codec, filepath.Base(row.Path))
			}
		default:
			return fmt.Errorf("M2TS cannot preserve selected %s track from %s; deselect it or use MKV", row.Track.Kind, filepath.Base(row.Path))
		}
	}
	if video == 0 {
		return fmt.Errorf("M2TS requires at least one selected Blu-ray-compatible video stream")
	}
	if video != 1 {
		return fmt.Errorf("M2TS requires exactly one selected video stream; found %d", video)
	}
	return nil
}

// Use the selected movie's duration so the final OGM chapter ends at the end
// of the film (rather than one second after the last chapter begins).
func mergerSelectedDuration(ctx context.Context, ffprobe string, streams []mergerStream) int64 {
	for _, item := range streams {
		if item.Track.Kind != "video" { continue }
		if item.BluRay {
			_, playlist, err := guiBlurayPlaylist(ctx, item.Path, item.BluRayPlaylist)
			if err == nil && playlist.DurationTicks > 0 {
				return playlist.DurationTicks * 1000 / 45000
			}
			continue
		}
		args := []string{"-v", "error"}
		if item.Track.DVDTitle > 0 {
			args = appendDesktopDVDInput(args, item.Track.DVDTitle, item.Path)
		} else {
			args = append(args, "-i", item.Path)
		}
		// FFprobe input options differ from ffmpeg's; use a simpler
		// standalone probe for ordinary file media.
		if item.Track.DVDTitle > 0 { continue }
		args = append(args, "-show_entries", "format=duration", "-of", "json")
		data, err := runMergerCommand(ctx, ffprobe, args...)
		if err != nil { continue }
		var result struct { Format struct { Duration string `json:"duration"` } `json:"format"` }
		if err = json.Unmarshal(data, &result); err != nil { continue }
		seconds, err := strconv.ParseFloat(result.Format.Duration, 64)
		if err == nil && seconds > 0 && seconds < 1000000 {
			return int64(seconds * 1000)
		}
	}
	return 0
}
