//go:build windows || linux

package main

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

// The desktop GUI deliberately opts into Blu-ray before routing sources here.
// A .iso extension alone does not distinguish a DVD image from a Blu-ray.
func guiBlurayPlaylists(ctx context.Context, path string) ([]titleInfo, error) {
	source, err := resolveBluraySource(path)
	if err != nil {
		return nil, err
	}
	playlists, err := discoverBlurayForSource(ctx, source)
	if err != nil {
		return nil, err
	}
	if len(playlists) == 0 {
		return nil, errors.New("Blu-ray contains no readable playlists")
	}
	titles := make([]titleInfo, 0, len(playlists))
	for _, playlist := range playlists {
		titles = append(titles, titleInfo{Number: playlist.Number, Duration: playlist.Duration()})
	}
	return titles, nil
}

func guiBlurayPlaylist(ctx context.Context, path string, playlistID int) (bluraySource, blurayPlaylist, error) {
	source, err := resolveBluraySource(path)
	if err != nil {
		return bluraySource{}, blurayPlaylist{}, err
	}
	playlists, err := discoverBlurayForSource(ctx, source)
	if err != nil {
		return bluraySource{}, blurayPlaylist{}, err
	}
	playlist, err := selectBlurayPlaylist(playlists, fmt.Sprintf("%05d", playlistID))
	return source, playlist, err
}

func guiBlurayTracks(ctx context.Context, path string, playlistID int) ([]trackOption, string, error) {
	source, playlist, err := guiBlurayPlaylist(ctx, path, playlistID)
	if err != nil {
		return nil, "", err
	}
	tools, err := resolveBlurayTools(ctx)
	if err != nil {
		return nil, "", err
	}
	streams, err := probeBlurayPlaylist(ctx, tools, source, playlist)
	if err != nil {
		return nil, "", err
	}
	options := make([]trackOption, 0, len(streams))
	var details strings.Builder
	fmt.Fprintf(&details, "Blu-ray playlist %05d\nDuration: %s\nChapters: %d\n\n", playlist.Number, playlist.Duration(), len(playlist.Chapters))
	for _, stream := range streams {
		options = append(options, trackOption{Index: stream.Index, Kind: stream.Kind, Codec: stream.Codec, Language: stream.Language, Title: bluRayStreamDisplayTitle(stream), Channels: stream.Channels})
		fmt.Fprintf(&details, "%s #%d: %s", stream.Kind, stream.Index, bluRayStreamDisplayTitle(stream))
		if validBlurayLanguage(stream.Language) { fmt.Fprintf(&details, " [%s]", stream.Language) }
		details.WriteString("\n")
	}
	details.WriteString("\nBlu-ray LPCM tracks are converted to lossless FLAC when remuxed. Other compatible tracks are copied. Blu-ray decryption and real physical drives are still experimental.\n")
	return options, details.String(), nil
}

func guiBlurayOutputPath(source bluraySource, playlistID int, dir string) string {
	stem := sanitizeFilename(source.BaseName)
	if stem == "" {
		stem = "Blu-ray"
	}
	return filepath.Join(dir, fmt.Sprintf("%s-bluray-%05d.mkv", stem, playlistID))
}

func guiBlurayRemux(ctx context.Context, path string, playlistID int, outputDir string, indexes []int, chapters bool) (string, error) {
	if err := validateOutputDir(outputDir); err != nil {
		return "", err
	}
	source, playlist, err := guiBlurayPlaylist(ctx, path, playlistID)
	if err != nil {
		return "", err
	}
	tools, err := resolveBlurayTools(ctx)
	if err != nil {
		return "", err
	}
	final := guiBlurayOutputPath(source, playlist.Number, outputDir)
	if err := remuxBlurayPlaylist(ctx, tools, source, playlist, final, indexes, chapters); err != nil {
		return "", err
	}
	return final, nil
}
