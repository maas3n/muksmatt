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

const (
	blurayVideoElementary = "elementary"
	blurayVideoM2TS       = "m2ts"
	blurayVideoBoth       = "both"
)

// Elementary outputs retain their codec-specific extension. An M2TS option
// *muxes* the copied Blu-ray video stream into 192-byte transport packets;
// it does not transcode, rename an elementary stream, or create an MKV.
func blurayDemuxFormats(stream blurayStream, videoMode string) ([]demuxFormat, error) {
	switch videoMode {
	case blurayVideoElementary, blurayVideoM2TS, blurayVideoBoth:
	default:
		return nil, fmt.Errorf("unsupported Blu-ray video export mode %q", videoMode)
	}
	if strings.EqualFold(stream.Codec, "pcm_bluray") {
		return []demuxFormat{{Extension: "flac", Muxer: "flac"}}, nil
	}
	elementary, err := streamDemuxFormat(stream.Codec, "mpeg2")
	if err != nil {
		return nil, err
	}
	if stream.Kind != "video" || videoMode == blurayVideoElementary {
		return []demuxFormat{elementary}, nil
	}
	switch strings.ToLower(stream.Codec) {
	case "h264", "hevc", "mpeg2video":
		// MPEG-TS supports these Blu-ray video codecs via stream copy.
	default:
		return nil, fmt.Errorf("Blu-ray M2TS export does not support video codec %s", stream.Codec)
	}
	transport := demuxFormat{Extension: "m2ts", Muxer: "mpegts"}
	if videoMode == blurayVideoM2TS {
		return []demuxFormat{transport}, nil
	}
	return []demuxFormat{elementary, transport}, nil
}

func blurayDemuxOutputArgs(args []string, stream blurayStream, format demuxFormat, path string) []string {
	codec := "copy"
	if strings.EqualFold(stream.Codec, "pcm_bluray") {
		codec = "flac"
	}
	args = append(args, "-map", fmt.Sprintf("0:%d", stream.Index), "-c", codec, "-map_metadata", "-1", "-map_chapters", "-1")
	if format.Filter != "" {
		args = append(args, "-bsf:v", format.Filter)
	}
	if format.Muxer == "mpegts" && format.Extension == "m2ts" {
		args = append(args, "-mpegts_m2ts_mode", "1")
	}
	if format.Muxer == "vob" {
		args = append(args, "-muxdelay", "0", "-preload", "0")
	}
	return append(args, "-f", format.Muxer, path)
}

// The existing REMUX/DEMUX tab and CLI continue to use elementary extraction.
func guiBlurayDemux(ctx context.Context, path string, playlistID int, outputDir string, indexes []int, chapters bool) (string, error) {
	return guiBlurayDemuxWithVideoMode(ctx, path, playlistID, outputDir, indexes, chapters, blurayVideoElementary)
}

// Demux the selected Blu-ray playlist directly via libbluray, never via a
// temporary MKV or any DVD-only input flags.
func guiBlurayDemuxWithVideoMode(ctx context.Context, path string, playlistID int, outputDir string, indexes []int, chapters bool, videoMode string) (final string, err error) {
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
		formats, formatErr := blurayDemuxFormats(stream, videoMode)
		if formatErr != nil {
			return "", formatErr
		}
		for _, format := range formats {
			exports = append(exports, export{stream: stream, format: format})
		}
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
		args = blurayDemuxOutputArgs(args, item.stream, item.format, path)
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
