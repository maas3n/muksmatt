//go:build windows || linux

package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// This source model is deliberately separate from the stable DVD engine.
// Recognizing an ISO by extension does not mean it has been verified as Blu-ray.
type bluraySourceKind uint8

const (
	bluraySourceDirectory bluraySourceKind = iota + 1
	bluraySourceISOCandidate
)

type bluraySource struct {
	Kind        bluraySourceKind
	Input       string
	Label       string
	BaseName    string
	PlaylistDir string // folder sources only; ISO requires a future libbluray reader
}

func resolveBluraySource(raw string) (bluraySource, error) {
	raw = strings.TrimSpace(strings.Trim(raw, "\""))
	if raw == "" {
		return bluraySource{}, errors.New("choose a Blu-ray disc root, BDMV folder or ISO file")
	}
	path := filepath.Clean(raw)
	st, err := os.Stat(path)
	if err != nil {
		return bluraySource{}, fmt.Errorf("Blu-ray source: %w", err)
	}
	if !st.IsDir() {
		if !st.Mode().IsRegular() || !strings.EqualFold(filepath.Ext(path), ".iso") {
			return bluraySource{}, errors.New("Blu-ray source must be a BDMV directory or ISO file")
		}
		return bluraySource{Kind: bluraySourceISOCandidate, Input: path,
			Label: filepath.Base(path), BaseName: strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))}, nil
	}
	bdmv, root := path, filepath.Dir(path)
	if !strings.EqualFold(filepath.Base(path), "BDMV") {
		bdmv, root = blurayChildDirFold(path, "BDMV"), path
	}
	if bdmv == "" {
		return bluraySource{}, errors.New("selected folder does not contain a BDMV directory")
	}
	playlistDir := blurayChildDirFold(bdmv, "PLAYLIST")
	if playlistDir == "" {
		return bluraySource{}, errors.New("BDMV directory does not contain PLAYLIST")
	}
	return bluraySource{Kind: bluraySourceDirectory, Input: root, Label: filepath.Base(root),
		BaseName: filepath.Base(root), PlaylistDir: playlistDir}, nil
}

func blurayChildDirFold(path, name string) string {
	entries, err := os.ReadDir(path)
	if err != nil {
		return ""
	}
	for _, entry := range entries {
		if entry.IsDir() && strings.EqualFold(entry.Name(), name) {
			return filepath.Join(path, entry.Name())
		}
	}
	return ""
}

type blurayChapter struct {
	Number     int
	StartTicks int64 // 45 kHz Blu-ray ticks
	EndTicks   int64
}

func (c blurayChapter) Start() time.Duration {
	return time.Duration(c.StartTicks * int64(time.Second) / 45000)
}
func (c blurayChapter) Duration() time.Duration {
	return time.Duration((c.EndTicks - c.StartTicks) * int64(time.Second) / 45000)
}

type blurayPlaylist struct {
	Number        int
	File          string
	DurationTicks int64
	Chapters      []blurayChapter
	ClipCount     int
}

func (p blurayPlaylist) Duration() time.Duration {
	return time.Duration(p.DurationTicks * int64(time.Second) / 45000)
}

// Discover navigation metadata only. The encrypted ISO and optical-disc read
// path is intentionally deferred until libbluray is integrated.
func discoverBlurayPlaylists(source bluraySource) ([]blurayPlaylist, error) {
	if source.Kind != bluraySourceDirectory || source.PlaylistDir == "" {
		return nil, errors.New("Blu-ray ISO and drive playlist scanning requires the future libbluray reader")
	}
	entries, err := os.ReadDir(source.PlaylistDir)
	if err != nil {
		return nil, fmt.Errorf("read Blu-ray PLAYLIST: %w", err)
	}
	var playlists []blurayPlaylist
	seen := map[int]bool{}
	for _, entry := range entries {
		if !entry.Type().IsRegular() {
			continue
		}
		name := entry.Name()
		if !strings.EqualFold(filepath.Ext(name), ".mpls") {
			continue
		}
		stem := name[:len(name)-len(filepath.Ext(name))]
		if len(stem) != 5 {
			continue
		}
		number, err := strconv.Atoi(stem)
		if err != nil || number < 0 || number > 99999 {
			continue
		}
		if seen[number] {
			return nil, fmt.Errorf("duplicate Blu-ray playlist %05d", number)
		}
		seen[number] = true
		p, err := parseBlurayMPLSFile(filepath.Join(source.PlaylistDir, name))
		if err != nil {
			return nil, fmt.Errorf("playlist %s: %w", name, err)
		}
		p.Number, p.File = number, filepath.Join(source.PlaylistDir, name)
		playlists = append(playlists, p)
	}
	if len(playlists) == 0 {
		return nil, errors.New("no readable MPLS playlists found")
	}
	sort.Slice(playlists, func(i, j int) bool { return playlists[i].Number < playlists[j].Number })
	return playlists, nil
}

// Default: longest playlist, with deterministic lowest-ID tie breaking.
// Explicit playlist selection cannot escape the discovered MPLS list.
func selectBlurayPlaylist(playlists []blurayPlaylist, requested string) (blurayPlaylist, error) {
	if len(playlists) == 0 {
		return blurayPlaylist{}, errors.New("no Blu-ray playlists")
	}
	requested = strings.TrimSpace(requested)
	if requested != "" {
		if strings.ContainsAny(requested, "/\\") {
			return blurayPlaylist{}, errors.New("playlist must be a number, not a path")
		}
		if strings.EqualFold(filepath.Ext(requested), ".mpls") {
			requested = requested[:len(requested)-5]
		}
		if len(requested) < 1 || len(requested) > 5 {
			return blurayPlaylist{}, errors.New("playlist must have 1 to 5 decimal digits")
		}
		for _, ch := range requested {
			if ch < '0' || ch > '9' {
				return blurayPlaylist{}, errors.New("playlist must contain digits only")
			}
		}
		n, err := strconv.Atoi(requested)
		if err != nil {
			return blurayPlaylist{}, err
		}
		for _, p := range playlists {
			if p.Number == n {
				return p, nil
			}
		}
		return blurayPlaylist{}, fmt.Errorf("playlist %05d not found", n)
	}
	best := playlists[0]
	for _, p := range playlists[1:] {
		if p.DurationTicks > best.DurationTicks ||
			(p.DurationTicks == best.DurationTicks && p.Number < best.Number) {
			best = p
		}
	}
	return best, nil
}
