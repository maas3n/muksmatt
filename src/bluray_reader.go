//go:build windows || linux

package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// The libbluray protocol exists only in FFmpeg builds configured with libbluray.
// Unlike a raw M2TS file, this URL follows the selected Blu-ray MPLS playlist.
func blurayProtocolURL(source bluraySource) (string, error) {
	if source.Kind != bluraySourceDirectory || source.PlaylistDir == "" {
		return "", errors.New("Blu-ray ISO and raw optical device reading is not yet supported; select a mounted BDMV disc root")
	}
	abs, err := filepath.Abs(source.Input)
	if err != nil {
		return "", err
	}
	return "bluray:" + filepath.ToSlash(abs), nil
}

func hasBlurayProtocol(listing []byte) bool {
	for _, line := range strings.Split(string(listing), "\n") {
		if strings.TrimSpace(line) == "bluray" {
			return true
		}
	}
	return false
}

func checkBlurayTools(ctx context.Context, tools toolPaths) error {
	for _, bin := range []string{tools.ffmpeg, tools.ffprobe} {
		if bin == "" {
			return errors.New("FFmpeg and FFprobe are required to read Blu-ray playlists")
		}
		protocols, err := runMergerCommand(ctx, bin, "-protocols")
		if err != nil {
			return fmt.Errorf("check libbluray protocol in %s: %w", filepath.Base(bin), err)
		}
		if !hasBlurayProtocol(protocols) {
			return fmt.Errorf("%s was built without the libbluray bluray: protocol; install a libbluray-enabled FFmpeg/FFprobe build", filepath.Base(bin))
		}
	}
	return nil
}

// Choose a single FFmpeg/FFprobe pair: never inadvertently combine an
// unrelated system ffprobe with a bundled ffmpeg build.
func resolveBlurayTools(ctx context.Context) (toolPaths, error) {
	ffmpeg, a := exec.LookPath("ffmpeg")
	ffprobe, b := exec.LookPath("ffprobe")
	if a == nil && b == nil {
		candidate := toolPaths{ffmpeg: ffmpeg, ffprobe: ffprobe}
		if err := checkBlurayTools(ctx, candidate); err == nil {
			return candidate, nil
		}
	}
	fallback, err := desktopCLITools(ctx, false)
	if err == nil {
		if err = checkBlurayTools(ctx, fallback); err == nil {
			return fallback, nil
		}
	}
	if err == nil {
		err = errors.New("no libbluray-capable FFmpeg/FFprobe pair found")
	}
	return toolPaths{}, fmt.Errorf("Blu-ray reading unavailable: %w", err)
}

func blurayInputArgs(source bluraySource, playlist blurayPlaylist) ([]string, error) {
	url, err := blurayProtocolURL(source)
	if err != nil {
		return nil, err
	}
	if playlist.Number < 0 || playlist.Number > 99999 {
		return nil, errors.New("invalid selected Blu-ray playlist number")
	}
	// Deliberately do NOT apply the DVD-only 100M/100M/+genpts policy.
	return []string{"-playlist", strconv.Itoa(playlist.Number), "-i", url}, nil
}

type blurayStream struct {
	Index int
	Kind  string
	Codec string
}

func parseBlurayProbe(data []byte) ([]blurayStream, error) {
	var result ffprobeResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("decode Blu-ray stream inventory: %w", err)
	}
	var selected []blurayStream
	seen := map[int]bool{}
	for _, stream := range result.Streams {
		switch stream.CodecType {
		case "video", "audio", "subtitle":
		default:
			continue
		}
		if stream.Index < 0 || seen[stream.Index] || stream.CodecName == "" {
			return nil, errors.New("Blu-ray probe contains invalid or duplicate stream indexes/codecs")
		}
		seen[stream.Index] = true
		selected = append(selected, blurayStream{Index: stream.Index, Kind: stream.CodecType, Codec: stream.CodecName})
	}
	if len(selected) == 0 {
		return nil, errors.New("no readable video/audio/subtitle streams; media may be encrypted or libaacs/libbdplus is not configured")
	}
	return selected, nil
}

func probeBlurayPlaylist(ctx context.Context, tools toolPaths, source bluraySource, playlist blurayPlaylist) ([]blurayStream, error) {
	if err := checkBlurayTools(ctx, tools); err != nil {
		return nil, err
	}
	input, err := blurayInputArgs(source, playlist)
	if err != nil {
		return nil, err
	}
	args := []string{"-hide_banner", "-v", "error", "-show_streams", "-of", "json"}
	args = append(args, input...)
	data, err := runMergerCommand(ctx, tools.ffprobe, args...)
	if err != nil {
		return nil, fmt.Errorf("libbluray could not read playlist %05d (verify disc access and configured AACS/BD+ support): %w", playlist.Number, err)
	}
	return parseBlurayProbe(data)
}

func selectedBlurayStreams(all []blurayStream, selected []int) ([]blurayStream, error) {
	if len(all) == 0 {
		return nil, errors.New("no Blu-ray streams")
	}
	if selected == nil {
		return append([]blurayStream(nil), all...), nil
	}
	if len(selected) == 0 {
		return nil, errors.New("choose at least one Blu-ray stream")
	}
	wanted := map[int]bool{}
	for _, id := range selected {
		if id < 0 || wanted[id] {
			return nil, fmt.Errorf("invalid or duplicate Blu-ray stream index %d", id)
		}
		wanted[id] = true
	}
	var result []blurayStream
	for _, s := range all {
		if wanted[s.Index] {
			result = append(result, s)
			delete(wanted, s.Index)
		}
	}
	if len(wanted) > 0 {
		return nil, errors.New("a requested Blu-ray stream does not exist in the selected playlist")
	}
	return result, nil
}

// FFmetadata is supplied as a second input; -map_chapters 0 does not reliably
// reconstruct chapters from FFmpeg's bluray byte-stream protocol.
func blurayFFmetadata(playlist blurayPlaylist) []byte {
	var out strings.Builder
	out.WriteString(";FFMETADATA1\n")
	for _, ch := range playlist.Chapters {
		if ch.EndTicks <= ch.StartTicks || ch.StartTicks < 0 {
			continue
		}
		fmt.Fprintf(&out, "[CHAPTER]\nTIMEBASE=1/45000\nSTART=%d\nEND=%d\n", ch.StartTicks, ch.EndTicks)
	}
	return []byte(out.String())
}

func blurayRemuxArgs(source bluraySource, playlist blurayPlaylist, streams []blurayStream, chaptersFile, partial string) ([]string, error) {
	if len(streams) == 0 || partial == "" {
		return nil, errors.New("Blu-ray remux requires selected streams and an output path")
	}
	input, err := blurayInputArgs(source, playlist)
	if err != nil {
		return nil, err
	}
	args := append([]string{"-hide_banner", "-nostdin", "-v", "error", "-y"}, input...)
	if chaptersFile != "" {
		args = append(args, "-f", "ffmetadata", "-i", chaptersFile)
	}
	for _, stream := range streams {
		if stream.Index < 0 {
			return nil, errors.New("negative Blu-ray stream index")
		}
		args = append(args, "-map", fmt.Sprintf("0:%d", stream.Index))
	}
	args = append(args, "-map_metadata", "0", "-map_chapters")
	if chaptersFile != "" {
		args = append(args, "1")
	} else {
		args = append(args, "-1")
	}
	// Place -c copy before stream-specific overrides so pcm_bluray always
	// becomes lossless FLAC, and every other selected stream is stream-copied.
	args = append(args, "-c", "copy")
	audioIndex := 0
	for _, stream := range streams {
		if stream.Kind == "audio" {
			if strings.EqualFold(stream.Codec, "pcm_bluray") {
				args = append(args, fmt.Sprintf("-c:a:%d", audioIndex), "flac")
			}
			audioIndex++
		}
	}
	args = append(args, "-f", "matroska", partial)
	return args, nil
}

func verifyBlurayRemux(ctx context.Context, probe string, filename string, selected []blurayStream) error {
	data, err := runMergerCommand(ctx, probe, "-v", "error", "-show_streams", "-of", "json", filename)
	if err != nil {
		return fmt.Errorf("verify remuxed Blu-ray MKV: %w", err)
	}
	got, err := parseBlurayProbe(data)
	if err != nil {
		return err
	}
	if len(got) != len(selected) {
		return fmt.Errorf("Blu-ray remux stream count mismatch: got %d, expected %d", len(got), len(selected))
	}
	for i, stream := range selected {
		want := stream.Codec
		if strings.EqualFold(want, "pcm_bluray") {
			want = "flac"
		}
		if got[i].Kind != stream.Kind || !strings.EqualFold(got[i].Codec, want) {
			return fmt.Errorf("Blu-ray stream %d mismatch: got %s/%s, expected %s/%s",
				i, got[i].Kind, got[i].Codec, stream.Kind, want)
		}
	}
	return nil
}

// This first native input path works with accessible BDMV folders / mounted
// disc roots. ISO image and raw optical device support will use a dedicated
// libbluray input adapter, not a made-up filesystem path.
func remuxBlurayPlaylist(ctx context.Context, tools toolPaths, source bluraySource, playlist blurayPlaylist, output string, selected []int, chapters bool) error {
	if source.Kind != bluraySourceDirectory {
		return errors.New("raw ISO and optical Blu-ray inputs require the next libbluray reader stage")
	}
	if !strings.EqualFold(filepath.Ext(output), ".mkv") {
		return errors.New("Blu-ray output must end in .mkv")
	}
	if err := checkBlurayTools(ctx, tools); err != nil {
		return err
	}
	streams, err := probeBlurayPlaylist(ctx, tools, source, playlist)
	if err != nil {
		return err
	}
	streams, err = selectedBlurayStreams(streams, selected)
	if err != nil {
		return err
	}
	output, err = filepath.Abs(output)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(output), 0755); err != nil {
		return err
	}
	if _, err := os.Lstat(output); err == nil {
		return fmt.Errorf("output already exists: %s", output)
	} else if !os.IsNotExist(err) {
		return err
	}
	partial, err := reservePartialOutput(output)
	if err != nil {
		return err
	}
	defer os.Remove(partial)
	metadataFile := ""
	if chapters && len(playlist.Chapters) != 0 {
		temp, err := os.CreateTemp(filepath.Dir(output), ".muksmatt-bluray-chapters-*.ffmeta")
		if err != nil {
			return err
		}
		metadataFile = temp.Name()
		defer os.Remove(metadataFile)
		if _, err := temp.Write(blurayFFmetadata(playlist)); err != nil {
			_ = temp.Close()
			return err
		}
		if err := temp.Close(); err != nil {
			return err
		}
	}
	args, err := blurayRemuxArgs(source, playlist, streams, metadataFile, partial)
	if err != nil {
		return err
	}
	if _, err := runMergerCommand(ctx, tools.ffmpeg, args...); err != nil {
		return fmt.Errorf("libbluray Blu-ray remux failed (confirm disc read access/AACS/BD+ support): %w", err)
	}
	if err := verifyBlurayRemux(ctx, tools.ffprobe, partial, streams); err != nil {
		return err
	}
	return finalizeRemuxOutput(partial, output)
}
