//go:build windows || linux

package main

import (
	"context"
	"fmt"
	"strings"
)

// Probe a Blu-ray source as one selected playlist. This is explicitly invoked
// by the Blu-ray control in ADVANCED, never inferred from the .iso extension.
func probeMergerBluRay(ctx context.Context, path, kind string) ([]mergerStream, error) {
	if kind != "all" && kind != "video" && kind != "audio" && kind != "subtitle" {
		return nil, fmt.Errorf("unsupported Blu-ray stream filter %q", kind)
	}
	source, err := resolveBluraySource(path)
	if err != nil {
		return nil, err
	}
	playlists, err := discoverBlurayForSource(ctx, source)
	if err != nil {
		return nil, err
	}
	playlist, err := selectBlurayPlaylist(playlists, "")
	if err != nil {
		return nil, err
	}
	tools, err := resolveBlurayTools(ctx)
	if err != nil {
		return nil, err
	}
	streams, err := probeBlurayPlaylist(ctx, tools, source, playlist)
	if err != nil {
		return nil, err
	}
	var result []mergerStream
	for _, stream := range streams {
		if kind != "all" && stream.Kind != kind {
			continue
		}
		result = append(result, mergerStream{
			Path: source.Input,
			Track: trackOption{
				Index: stream.Index, Kind: stream.Kind, Codec: stream.Codec,
			},
			BluRay: true,
			BluRayPlaylist: playlist.Number,
		})
	}
	if kind == "all" && len(playlist.Chapters) > 0 {
		result = append(result, mergerStream{
			Path: source.Input,
			Track: trackOption{Index: -1, Kind: "chapters", Title: fmt.Sprintf("%d Blu-ray chapters (playlist %05d)", len(playlist.Chapters), playlist.Number)},
			BluRay: true,
			BluRayPlaylist: playlist.Number,
		})
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("Blu-ray playlist %05d contains no matching %s tracks", playlist.Number, kind)
	}
	return result, nil
}

func selectedBluRayMergerStreams(streams []mergerStream) bool {
	for _, stream := range streams {
		if stream.BluRay {
			return true
		}
	}
	return false
}

// Verify that every Blu-ray track from one path belongs to one MPLS playlist.
func mergerBluraySelections(streams []mergerStream) (map[string]int, error) {
	playlists := map[string]int{}
	for _, stream := range streams {
		if !stream.BluRay {
			continue
		}
		if previous, ok := playlists[stream.Path]; ok && previous != stream.BluRayPlaylist {
			return nil, fmt.Errorf("conflicting Blu-ray playlist selections for %s", stream.Path)
		}
		playlists[stream.Path] = stream.BluRayPlaylist
	}
	return playlists, nil
}

// The existing stream map in ADVANCED can include ordinary containers, DVDs
// and one or more Blu-ray sources. The final output audio index is used to
// force lossless FLAC specifically for each selected Blu-ray LPCM stream.
func mergerBlurayLPCMAudioIndexes(streams []mergerStream) []int {
	seen := map[string]bool{}
	audioIndex := 0
	var converted []int
	for _, stream := range streams {
		if stream.Track.Kind == "chapters" {
			continue
		}
		key := fmt.Sprintf("%s:%d", stream.Path, stream.Track.Index)
		if seen[key] {
			continue
		}
		seen[key] = true
		if stream.Track.Kind == "audio" {
			if stream.BluRay && strings.EqualFold(stream.Track.Codec, "pcm_bluray") {
				converted = append(converted, audioIndex)
			}
			audioIndex++
		}
	}
	return converted
}
