//go:build windows || linux

package main

import (
	"context"
	"errors"
	"fmt"
	"os"
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
				Language: stream.Language, Title: bluRayStreamDisplayTitle(stream), Channels: stream.Channels,
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

 
// Blu-ray DEMUX uses a direct libbluray reader for each selected playlist.
// Keep ordinary Advanced Merger inputs on their existing DVD/MKV demux path.
// Multiple Blu-ray selections get source-specific directories under a single
// new output root, which is removed on error or cancellation.
func demuxMergerBluRay(ctx context.Context, selected []mergerStream, outputDir string) (final string, err error) {
	type group struct {
		path string
		playlist int
		indexes []int
		chapters bool
	}
	var groups []*group
	byPath := map[string]*group{}
	for _, item := range selected {
		if !item.BluRay {
			return "", errors.New("DEMUX Blu-ray sources separately from ordinary DVD/MKV inputs; mixed MUX remains supported")
		}
		grp := byPath[item.Path]
		if grp == nil {
			grp = &group{path: item.Path, playlist: item.BluRayPlaylist}
			byPath[item.Path] = grp
			groups = append(groups, grp)
		}
		if grp.playlist != item.BluRayPlaylist {
			return "", errors.New("conflicting Blu-ray playlists for one input")
		}
		switch item.Track.Kind {
		case "chapters":
			grp.chapters = true
		case "video", "audio", "subtitle":
			grp.indexes = append(grp.indexes, item.Track.Index)
		default:
			return "", fmt.Errorf("unsupported Blu-ray DEMUX stream type: %s", item.Track.Kind)
		}
	}
	if len(groups) == 0 {
		return "", errors.New("select at least one Blu-ray source")
	}
	for _, group := range groups {
		if len(group.indexes) == 0 {
			return "", errors.New("choose at least one media stream from every Blu-ray source")
		}
	}
	if len(groups) == 1 {
		group := groups[0]
		return guiBlurayDemux(ctx, group.path, group.playlist, outputDir, group.indexes, group.chapters)
	}
	if err = validateOutputDir(outputDir); err != nil {
		return "", err
	}
	final, err = os.MkdirTemp(outputDir, "Advanced-Blu-ray-Demux-")
	if err != nil {
		return "", err
	}
	defer func() {
		if err != nil {
			_ = os.RemoveAll(final)
		}
	}()
	for _, group := range groups {
		if _, err = guiBlurayDemux(ctx, group.path, group.playlist, final, group.indexes, group.chapters); err != nil {
			return final, err
		}
	}
	return final, nil
}
