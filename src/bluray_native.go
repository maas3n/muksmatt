//go:build windows || linux

package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const blurayNativeProtocol = "MUKSMATT_BD_NAV_1"

// resolveBlurayNavigator searches the installed companion next to muKsMaTT,
// then PATH. A name match alone is not trusted: parsing requires its versioned
// protocol and validates each playlist/chapter before any FFmpeg command runs.
func resolveBlurayNavigator() (string, error) {
	name := "muksmatt-bluray-nav"
	if exe, err := os.Executable(); err == nil {
		sibling := filepath.Join(filepath.Dir(exe), name)
		if _, err := os.Stat(sibling); err == nil {
			return sibling, nil
		}
		if _, err := os.Stat(sibling + ".exe"); err == nil {
			return sibling + ".exe", nil
		}
	}
	if exe, err := exec.LookPath(name); err == nil {
		return exe, nil
	}
	return "", errors.New("Blu-ray ISO/device navigation requires the compiled libbluray companion (muksmatt-bluray-nav); consult docs/BLURAY_PHASE3.md")
}

func parseNativeBlurayNavigation(output []byte) ([]blurayPlaylist, error) {
	if len(output) > 8<<20 {
		return nil, errors.New("libbluray navigation output is oversized")
	}
	rows := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(rows) < 2 || strings.TrimSpace(rows[0]) != blurayNativeProtocol {
		return nil, errors.New("unsupported native libbluray navigation protocol")
	}
	playlists := make(map[int]blurayPlaylist)
	type nativeChapter struct{ start, end int64 }
	chapters := make(map[int][]nativeChapter)
	rawDuration := make(map[int]int64)
	for _, row := range rows[1:] {
		fields := strings.Split(strings.TrimSpace(row), "\t")
		if len(fields) == 0 || len(fields[0]) != 1 {
			return nil, errors.New("invalid native Blu-ray navigation record")
		}
		switch fields[0] {
		case "P":
			if len(fields) != 4 {
				return nil, errors.New("invalid native playlist record")
			}
			pid, err := parseNativeBlurayInt(fields[1], 99999)
			if err != nil { return nil, err }
			ticks, err := parseNativeBlurayTicks(fields[2])
			if err != nil || ticks <= 0 {
				return nil, errors.New("invalid native playlist duration")
			}
			clips, err := parseNativeBlurayInt(fields[3], 10000)
			if err != nil || clips <= 0 {
				return nil, errors.New("invalid native clip count")
			}
			if _, exists := playlists[pid]; exists {
				return nil, fmt.Errorf("duplicate native Blu-ray playlist %05d", pid)
			}
			playlists[pid] = blurayPlaylist{
				Number: pid, DurationTicks: ticks / 2, ClipCount: clips,
			}
			rawDuration[pid] = ticks
		case "C":
			if len(fields) != 4 {
				return nil, errors.New("invalid native chapter record")
			}
			pid, err := parseNativeBlurayInt(fields[1], 99999)
			if err != nil { return nil, err }
			start, err := parseNativeBlurayTicks(fields[2])
			if err != nil { return nil, err }
			end, err := parseNativeBlurayTicks(fields[3])
			if err != nil || end <= start {
				return nil, errors.New("invalid native chapter interval")
			}
			chapters[pid] = append(chapters[pid], nativeChapter{start: start, end: end})
			if len(chapters[pid]) > 10000 {
				return nil, errors.New("too many native Blu-ray chapters")
			}
		default:
			return nil, fmt.Errorf("unexpected native Blu-ray record %q", fields[0])
		}
	}
	if len(playlists) == 0 {
		return nil, errors.New("libbluray did not report playlists")
	}
	for pid := range chapters {
		if _, ok := playlists[pid]; !ok {
			return nil, fmt.Errorf("chapter belongs to unknown playlist %05d", pid)
		}
	}
	result := make([]blurayPlaylist, 0, len(playlists))
	for pid, p := range playlists {
		entries := chapters[pid]
		sort.Slice(entries, func(i,j int)bool{
			return entries[i].start < entries[j].start
		})
		for i, ch := range entries {
			// Validate in the original native 90 kHz timebase. Converting
			// to 45 kHz first would hide one-tick chapter overruns.
			if ch.start < 0 || ch.end > rawDuration[pid] || ch.end <= ch.start ||
				(i > 0 && ch.start < entries[i-1].end) {
				return nil, fmt.Errorf("invalid chapter timeline for playlist %05d", pid)
			}
			converted := blurayChapter{
				Number: i + 1, StartTicks: ch.start / 2, EndTicks: ch.end / 2,
			}
			if converted.EndTicks <= converted.StartTicks {
				return nil, fmt.Errorf("chapter shorter than one 45 kHz tick on playlist %05d", pid)
			}
			p.Chapters = append(p.Chapters, converted)
		}
		result = append(result, p)
	}
	sort.Slice(result, func(i, j int) bool {return result[i].Number < result[j].Number})
	return result, nil
}
func parseNativeBlurayInt(text string, max int) (int, error) {
	if text == "" || len(text) > 10 {return 0, errors.New("invalid Blu-ray navigation integer")}
	for _, ch := range text {
		if ch < '0' || ch > '9' {return 0, errors.New("invalid Blu-ray navigation integer")}
	}
	n, err := strconv.ParseUint(text, 10, 32)
	if err != nil || n > uint64(max) {
		return 0, errors.New("Blu-ray navigation integer out of range")
	}
	return int(n), nil
}
func parseNativeBlurayTicks(text string) (int64, error) {
	if text == "" || len(text) > 16 {return 0, errors.New("invalid Blu-ray timestamp")}
	for _, ch := range text {
		if ch < '0' || ch > '9' {return 0, errors.New("invalid Blu-ray timestamp")}
	}
	n, err := strconv.ParseUint(text, 10, 64)
	if err != nil || n > uint64(maxBlurayTicks)*2 {
		return 0, errors.New("Blu-ray navigation timestamp out of range")
	}
	return int64(n), nil
}
func discoverBlurayNative(ctx context.Context, source bluraySource) ([]blurayPlaylist, error) {
	if source.Kind != bluraySourceISOCandidate && source.Kind != bluraySourcePhysicalDrive {
		return nil, errors.New("native navigator expected an ISO image or optical drive")
	}
	helper, err := resolveBlurayNavigator()
	if err != nil {return nil, err}
	// No shell. Source is passed as one opaque argument, including paths with spaces.
	cmd := exec.CommandContext(ctx, helper, source.Input)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("libbluray cannot inspect %s: %w (check media access, libaacs/libbdplus and authorization)", source.Input, err)
	}
	return parseNativeBlurayNavigation(output)
}

func discoverBlurayForSource(ctx context.Context, source bluraySource) ([]blurayPlaylist, error) {
	switch source.Kind {
	case bluraySourceDirectory:
		return discoverBlurayPlaylists(source)
	case bluraySourceISOCandidate, bluraySourcePhysicalDrive:
		return discoverBlurayNative(ctx, source)
	default:
		return nil, errors.New("unrecognized Blu-ray source")
	}
}
