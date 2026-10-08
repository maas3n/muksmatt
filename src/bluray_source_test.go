//go:build windows || linux

package main

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// A two-clip playlist with a 60-second program timeline and a discontinuous
// 100-second clock reset in the second clip. The second chapter starts at 40s.
func syntheticBlurayMPLS(secondTicks int64, duplicate bool) []byte {
	b := make([]byte, 512)
	copy(b[:8], "MPLS0300")
	binary.BigEndian.PutUint32(b[8:], 64)
	binary.BigEndian.PutUint32(b[12:], 256)
	binary.BigEndian.PutUint32(b[64:], 54)
	binary.BigEndian.PutUint16(b[70:], 2)
	binary.BigEndian.PutUint16(b[74:], 20)
	binary.BigEndian.PutUint32(b[88:], 0)
	binary.BigEndian.PutUint32(b[92:], 40*45000)
	binary.BigEndian.PutUint16(b[96:], 20)
	binary.BigEndian.PutUint32(b[110:], 100*45000)
	binary.BigEndian.PutUint32(b[114:], uint32(100*45000+secondTicks))
	count := 2
	if duplicate { count = 3 }
	binary.BigEndian.PutUint32(b[256:], uint32(2+14*count))
	binary.BigEndian.PutUint16(b[260:], uint16(count))
	b[263] = 1
	binary.BigEndian.PutUint16(b[264:], 0)
	binary.BigEndian.PutUint32(b[266:], 0)
	b[277] = 1
	binary.BigEndian.PutUint16(b[278:], 1)
	binary.BigEndian.PutUint32(b[280:], 100*45000)
	if duplicate {
		b[291] = 1
		binary.BigEndian.PutUint16(b[292:], 1)
		binary.BigEndian.PutUint32(b[294:], 100*45000)
	}
	return b
}

func TestBlurayMPLSTimelineAndChapters(t *testing.T) {
	p, err := parseBlurayMPLS(syntheticBlurayMPLS(20*45000, false))
	if err != nil { t.Fatal(err) }
	if p.ClipCount != 2 || p.Duration() != 60*time.Second || len(p.Chapters) != 2 {
		t.Fatalf("unexpected playlist: %+v", p)
	}
	if p.Chapters[0].Start() != 0 || p.Chapters[0].Duration() != 40*time.Second {
		t.Fatalf("first chapter: %+v", p.Chapters[0])
	}
	if p.Chapters[1].Start() != 40*time.Second || p.Chapters[1].Duration() != 20*time.Second {
		t.Fatalf("second chapter: %+v", p.Chapters[1])
	}
}

func TestBlurayMPLSMalformedInput(t *testing.T) {
	fixtures := map[string]func([]byte) []byte{
		"short": func(_ []byte) []byte { return []byte("MPLS") },
		"signature": func(b []byte) []byte { copy(b[:8], "INVALID!"); return b },
		"section bounds": func(b []byte) []byte { binary.BigEndian.PutUint32(b[12:], 600); return b },
		"mark count": func(b []byte) []byte { binary.BigEndian.PutUint16(b[260:], 99); return b },
		"playitem length": func(b []byte) []byte { binary.BigEndian.PutUint16(b[74:], 10); return b },
		"clock reversal": func(b []byte) []byte { binary.BigEndian.PutUint32(b[92:], 0); return b },
		"48hr limit": func(b []byte) []byte {
			binary.BigEndian.PutUint32(b[64:], 116)
			binary.BigEndian.PutUint16(b[70:], 5)
			for i := 0; i < 5; i++ {
				pos := 74 + i*22
				binary.BigEndian.PutUint16(b[pos:], 20)
				binary.BigEndian.PutUint32(b[pos+14:], 0)
				binary.BigEndian.PutUint32(b[pos+18:], 0x7fffffff)
			}
			return b
		},
	}
	for name, mutate := range fixtures {
		t.Run(name, func(t *testing.T) {
			if _, err := parseBlurayMPLS(mutate(syntheticBlurayMPLS(20*45000, false))); err == nil {
				t.Fatal("accepted malformed MPLS")
			}
		})
	}
}

func TestBlurayDuplicateMarks(t *testing.T) {
	p, err := parseBlurayMPLS(syntheticBlurayMPLS(20*45000, true))
	if err != nil { t.Fatal(err) }
	if len(p.Chapters) != 2 { t.Fatalf("expected 2 distinct chapters, got %d", len(p.Chapters)) }
}

func TestBlurayDirectoryAndPlaylistChoice(t *testing.T) {
	root := filepath.Join(t.TempDir(), "Feature")
	dir := filepath.Join(root, "bdmv", "playlist")
	if err := os.MkdirAll(dir, 0755); err != nil { t.Fatal(err) }
	for name, ticks := range map[string]int64{
		"00800.MPLS":30*45000, "00010.mpls":20*45000, "00002.mpls":20*45000,
	} {
		if err := os.WriteFile(filepath.Join(dir, name), syntheticBlurayMPLS(ticks, false), 0644); err != nil {
			t.Fatal(err)
		}
	}
	for _, input := range []string{root, filepath.Join(root, "bdmv")} {
		s, err := resolveBluraySource(input)
		if err != nil { t.Fatal(err) }
		if s.Kind != bluraySourceDirectory || s.Input != root || s.PlaylistDir != dir {
			t.Fatalf("unexpected source: %+v", s)
		}
		playlists, err := discoverBlurayPlaylists(s)
		if err != nil { t.Fatal(err) }
		if len(playlists) != 3 { t.Fatalf("got %d playlists", len(playlists)) }
		longest, err := selectBlurayPlaylist(playlists, "")
		if err != nil || longest.Number != 800 || longest.Duration() != 70*time.Second {
			t.Fatalf("longest: %+v %v", longest, err)
		}
		explicit, err := selectBlurayPlaylist(playlists, "00002.mpls")
		if err != nil || explicit.Number != 2 { t.Fatalf("explicit: %+v %v", explicit, err) }
		if _, err := selectBlurayPlaylist(playlists, "12345"); err == nil { t.Fatal("unknown playlist accepted") }
		if _, err := selectBlurayPlaylist(playlists, "../00800"); err == nil { t.Fatal("playlist path accepted") }
	}
}

func TestBlurayISOIsOnlyCandidate(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "disc.ISO")
	if err := os.WriteFile(path, []byte("not a real ISO"), 0644); err != nil { t.Fatal(err) }
	s, err := resolveBluraySource(path)
	if err != nil { t.Fatal(err) }
	if s.Kind != bluraySourceISOCandidate || s.PlaylistDir != "" { t.Fatalf("source: %+v", s) }
	if _, err := discoverBlurayPlaylists(s); err == nil || !strings.Contains(err.Error(), "future libbluray") {
		t.Fatal("ISO was incorrectly advertised as a scanned Blu-ray")
	}
	bad := filepath.Join(dir, "video.mkv")
	_ = os.WriteFile(bad, []byte("fixture"), 0644)
	if _, err := resolveBluraySource(bad); err == nil { t.Fatal("accepted MKV as Blu-ray") }
	if _, err := resolveBluraySource(dir); err == nil { t.Fatal("accepted ordinary folder") }
}

func TestBlurayMalformedPlaylistIsReported(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "BDMV", "PLAYLIST")
	if err := os.MkdirAll(dir, 0755); err != nil { t.Fatal(err) }
	_ = os.WriteFile(filepath.Join(dir, "00001.mpls"), syntheticBlurayMPLS(20*45000, false), 0644)
	_ = os.WriteFile(filepath.Join(dir, "00002.mpls"), []byte("garbage"), 0644)
	s, err := resolveBluraySource(root)
	if err != nil { t.Fatal(err) }
	if _, err := discoverBlurayPlaylists(s); err == nil {
		t.Fatal("corrupt navigation file silently ignored")
	}
}
