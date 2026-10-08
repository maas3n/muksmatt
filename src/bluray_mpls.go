//go:build windows || linux

package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"sort"
)

// MPLS files contain navigation data, not the decrypted movie bitstream.
const maxMPLSBytes int64 = 16 << 20
const maxBlurayTicks int64 = 45000 * 60 * 60 * 48

func parseBlurayMPLSFile(path string) (blurayPlaylist, error) {
	f, err := os.Open(path)
	if err != nil {
		return blurayPlaylist{}, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return blurayPlaylist{}, err
	}
	if info.Size() < 20 || info.Size() > maxMPLSBytes {
		return blurayPlaylist{}, fmt.Errorf("invalid MPLS size %d", info.Size())
	}
	b := make([]byte, info.Size())
	if _, err := f.ReadAt(b, 0); err != nil {
		return blurayPlaylist{}, err
	}
	return parseBlurayMPLS(b)
}

// A bounded parser for the MPLS PlayList and PlayListMark sections.
// Only main-angle PlayItems and type-1 chapter marks are modeled; no
// feature-playback, seamless branching or decryption is claimed here.
func parseBlurayMPLS(b []byte) (blurayPlaylist, error) {
	fail := func(msg string) (blurayPlaylist, error) {
		return blurayPlaylist{}, errors.New(msg)
	}
	if len(b) < 20 {
		return fail("truncated MPLS header")
	}
	switch string(b[:8]) {
	case "MPLS0100", "MPLS0200", "MPLS0240", "MPLS0300":
	default:
		return fail("unsupported MPLS signature")
	}
	playlistOff := uint64(binary.BigEndian.Uint32(b[8:12]))
	marksOff := uint64(binary.BigEndian.Uint32(b[12:16]))
	bounds := func(off, n uint64) bool {
		return off <= uint64(len(b)) && n <= uint64(len(b))-off
	}
	u16 := func(off uint64) uint16 { return binary.BigEndian.Uint16(b[int(off):int(off+2)]) }
	u32 := func(off uint64) uint32 { return binary.BigEndian.Uint32(b[int(off):int(off+4)]) }
	time45 := func(off uint64) int64 { return int64(u32(off) & 0x7fffffff) }

	if !bounds(playlistOff, 10) || !bounds(marksOff, 6) {
		return fail("invalid MPLS section offsets")
	}
	playlistSize := uint64(u32(playlistOff)) + 4
	markSize := uint64(u32(marksOff)) + 4
	if playlistSize < 10 || markSize < 6 ||
		!bounds(playlistOff, playlistSize) || !bounds(marksOff, markSize) {
		return fail("truncated MPLS sections")
	}
	playlistEnd, marksEnd := playlistOff+playlistSize, marksOff+markSize
	type clip struct{ in, out, relative int64 }
	itemCount := int(u16(playlistOff+6))
	if itemCount == 0 {
		return fail("MPLS contains no PlayItems")
	}
	items := make([]clip, 0, itemCount)
	pos := playlistOff + 10
	var total int64
	for i := 0; i < itemCount; i++ {
		if pos+22 > playlistEnd {
			return fail("truncated MPLS PlayItem")
		}
		itemLen := uint64(u16(pos))
		if itemLen < 20 || pos+2+itemLen > playlistEnd {
			return fail("invalid MPLS PlayItem length")
		}
		in, out := time45(pos+14), time45(pos+18)
		if out <= in {
			return fail("non-positive MPLS PlayItem duration")
		}
		interval := out - in
		if total+interval > maxBlurayTicks {
			return fail("MPLS duration exceeds 48 hours")
		}
		items = append(items, clip{in: in, out: out, relative: total})
		total += interval
		pos += 2 + itemLen
	}
	markCount := int(u16(marksOff+4))
	pos = marksOff + 6
	if uint64(markCount) > (marksEnd-pos)/14 {
		return fail("truncated MPLS PlayListMark entries")
	}
	seen := make(map[int64]bool, markCount)
	starts := make([]int64, 0, markCount)
	for i := 0; i < markCount; i++ {
		markType := b[int(pos+1)]
		clipIndex := int(u16(pos + 2))
		markTime := time45(pos + 4)
		if markType == 1 && clipIndex < len(items) {
			c := items[clipIndex]
			// Never promote an out-of-clip timestamp to a chapter.
			if markTime >= c.in && markTime < c.out {
				relative := c.relative + markTime - c.in
				// Ignore invalid and conventional last-second marks.
				if relative >= 0 && total-relative > 45000 && !seen[relative] {
					starts = append(starts, relative)
					seen[relative] = true
				}
			}
		}
		pos += 14
	}
	sort.Slice(starts, func(i, j int) bool { return starts[i] < starts[j] })
	chapters := make([]blurayChapter, 0, len(starts))
	for i, start := range starts {
		end := total
		if i+1 < len(starts) {
			end = starts[i+1]
		}
		if end > start {
			chapters = append(chapters, blurayChapter{
				Number: len(chapters) + 1, StartTicks: start, EndTicks: end,
			})
		}
	}
	return blurayPlaylist{
		DurationTicks: total, Chapters: chapters, ClipCount: len(items),
	}, nil
}
