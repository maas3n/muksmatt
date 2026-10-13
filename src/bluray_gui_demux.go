//go:build windows || linux

package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Demux a selected Blu-ray playlist directly through the same libbluray
// transport used by the desktop CLI. Never apply DVD-only input flags or
// produce a temporary source.mkv. Unsupported codecs fail before files appear.
func guiBlurayDemux(ctx context.Context, path string, playlistID int, outputDir string, indexes []int, chapters bool) (final string, err error) {
	if err = validateOutputDir(outputDir); err != nil {
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
	available, err := probeBlurayPlaylist(ctx, tools, source, playlist)
	if err != nil {
		return "", err
	}
	selected, err := selectedBlurayStreams(available, indexes)
	if err != nil {
		return "", err
	}
	type export struct {
		stream blurayStream
		format demuxFormat
	}
	exports := make([]export, 0, len(selected))
	for _, stream := range selected {
		var format demuxFormat
		if strings.EqualFold(stream.Codec, "pcm_bluray") {
			format = demuxFormat{Extension: "flac", Muxer: "flac"}
		} else {
			format, err = streamDemuxFormat(stream.Codec, "mpeg2")
			if err != nil {
				return "", err
			}
		}
		exports = append(exports, export{stream: stream, format: format})
	}
	stem := sanitizeFilename(source.BaseName)
	if stem == "" {
		stem = "Blu-ray"
	}
	final, err = os.MkdirTemp(outputDir, fmt.Sprintf("%s-playlist-%05d-demux-", stem, playlist.Number))
	if err != nil {
		return "", err
	}
	defer func() {
		if err != nil {
			_ = os.RemoveAll(final)
		}
	}()

	args := []string{"-hide_banner", "-v", "error", "-nostdin", "-n"}
	input, err := blurayInputArgs(source, playlist)
	if err != nil {
		return "", err
	}
	args = append(args, input...)
	for _, item := range exports {
		path := filepath.Join(final, fmt.Sprintf("track-%02d.%s", item.stream.Index, item.format.Extension))
		codec := "copy"
		if strings.EqualFold(item.stream.Codec, "pcm_bluray") {
			codec = "flac"
		}
		args = append(args, "-map", fmt.Sprintf("0:%d", item.stream.Index), "-c", codec, "-map_metadata", "-1", "-map_chapters", "-1")
		if item.format.Filter != "" {
			args = append(args, "-bsf:v", item.format.Filter)
		}
		if item.format.Muxer == "vob" {
			args = append(args, "-muxdelay", "0", "-preload", "0")
		}
		args = append(args, "-f", item.format.Muxer, path)
	}
	if _, err = runMergerCommand(ctx, tools.ffmpeg, args...); err != nil {
		return final, fmt.Errorf("Blu-ray stream extraction failed: %w", err)
	}
	for _, item := range exports {
		path := filepath.Join(final, fmt.Sprintf("track-%02d.%s", item.stream.Index, item.format.Extension))
		if err = validateAndSyncOutput(path); err != nil {
			return final, err
		}
	}
	if chapters && len(playlist.Chapters) > 0 {
		var b strings.Builder
		for i, chapter := range playlist.Chapters {
			fmt.Fprintf(&b, "CHAPTER%02d=%s\nCHAPTER%02dNAME=Chapter %d\n", i+1, chapterTimestamp(chapter.Start().Milliseconds()), i+1, i+1)
		}
		chapterPath := filepath.Join(final, "Chapters.txt")
		if err = os.WriteFile(chapterPath, []byte(b.String()), 0644); err != nil {
			return final, err
		}
		if err = validateAndSyncOutput(chapterPath); err != nil {
			return final, err
		}
	}
	if len(exports) == 0 {
		return final, errors.New("no Blu-ray streams selected")
	}
	return final, nil
}
