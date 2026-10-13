//go:build windows || linux

package main

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// The Blu-ray libbluray navigator supplies ISO 639-2 language codes indexed
// by MPEG-TS PID. FFprobe's bluray: demuxer often omits those language tags.
// A missing PID or a conflicting language must never be guessed.
type blurayLanguageKey struct {
	kind string
	pid  uint16
}

const blurayLanguageProtocol = "MUKSMATT_BD_LANG_1"

func validBlurayLanguage(lang string) bool {
	if len(lang) != 3 || strings.EqualFold(lang, "und") {
		return false
	}
	for _, ch := range lang {
		if ch < 'a' || ch > 'z' {
			return false
		}
	}
	return true
}

func parseNativeBlurayLanguages(data []byte) (map[blurayLanguageKey]string, error) {
	if len(data) > 1<<20 {
		return nil, errors.New("Blu-ray stream language inventory is oversized")
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) < 1 || strings.TrimSpace(lines[0]) != blurayLanguageProtocol {
		return nil, errors.New("unsupported Blu-ray stream language protocol")
	}
	result := make(map[blurayLanguageKey]string)
	ambiguous := make(map[blurayLanguageKey]bool)
	for _, line := range lines[1:] {
		fields := strings.Split(strings.TrimSpace(line), "\t")
		if len(fields) != 4 || fields[0] != "S" || (fields[1] != "a" && fields[1] != "s") {
			return nil, fmt.Errorf("invalid Blu-ray stream language record")
		}
		pid, err := strconv.ParseUint(fields[2], 10, 13)
		lang := strings.ToLower(fields[3])
		if err != nil || pid == 0 || pid >= 8192 || len(lang) != 3 {
			return nil, fmt.Errorf("invalid Blu-ray PID or language record")
		}
		if lang == "und" { continue }
		if !validBlurayLanguage(lang) {
			return nil, errors.New("invalid Blu-ray stream language code")
		}
		kind := "audio"
		if fields[1] == "s" {
			kind = "subtitle"
		}
		key := blurayLanguageKey{kind: kind, pid: uint16(pid)}
		if prev, present := result[key]; present && prev != lang {
			delete(result, key)
			ambiguous[key] = true
		}
		if !ambiguous[key] {
			result[key] = lang
		}
	}
	return result, nil
}

func nativeBlurayLanguages(ctx context.Context, source bluraySource, playlist int) (map[blurayLanguageKey]string, error) {
	navigator, err := resolveBlurayNavigator()
	if err != nil {
		return nil, err
	}
	if playlist < 0 || playlist > 99999 {
		return nil, errors.New("Blu-ray playlist number outside valid range")
	}
	cmd := exec.CommandContext(ctx, navigator, "--stream-languages", strconv.Itoa(playlist), source.Input)
	data, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("read libbluray languages for playlist %05d: %w", playlist, err)
	}
	return parseNativeBlurayLanguages(data)
}

func addBlurayLanguages(streams []blurayStream, sourceLanguages map[blurayLanguageKey]string) {
	for i := range streams {
		stream := &streams[i]
		if stream.Language != "" || stream.PID == 0 {
			continue
		}
		if lang := sourceLanguages[blurayLanguageKey{kind: stream.Kind, pid: stream.PID}]; validBlurayLanguage(lang) {
			stream.Language = lang
		}
	}
}

// A track's title in a Blu-ray can be absent. Generate only codec-based
// labels from verified probe fields; do not infer spoken language or content.
func bluRayStreamDisplayTitle(stream blurayStream) string {
	if stream.Title != "" {
		return stream.Title
	}
	var title string
	switch stream.Codec {
	case "hdmv_pgs_subtitle":
		title = "PGS subtitles"
	case "pcm_bluray":
		title = "LPCM (lossless FLAC in MKV)"
	default:
		title = strings.ToUpper(strings.ReplaceAll(stream.Codec, "_", " "))
		if stream.Profile != "" && !strings.EqualFold(stream.Profile, "unknown") {
			title = stream.Profile
		}
	}
	if stream.Kind == "audio" && stream.Channels > 0 {
		title += fmt.Sprintf(" (%d channels)", stream.Channels)
	}
	return title
}

func blurayMappedStreamMetadataArgs(streams []blurayStream) []string {
	var args []string
	audioIndex, subtitleIndex := 0, 0
	for _, stream := range streams {
		var spec string
		switch stream.Kind {
		case "audio":
			spec = fmt.Sprintf("a:%d", audioIndex)
			audioIndex++
		case "subtitle":
			spec = fmt.Sprintf("s:%d", subtitleIndex)
			subtitleIndex++
		default:
			continue
		}
		if validBlurayLanguage(stream.Language) {
			args = append(args, "-metadata:s:"+spec, "language="+stream.Language)
		}
		if label := bluRayStreamDisplayTitle(stream); label != "" {
			args = append(args, "-metadata:s:"+spec, "title="+label)
		}
	}
	return args
}
