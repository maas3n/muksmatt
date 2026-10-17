//go:build windows || linux

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var ogmChapterLine = regexp.MustCompile(`(?i)^CHAPTER([0-9]{1,5})(NAME)?=(.*)$`)

type ogmChapter struct {
	Start int64 // milliseconds
	Title string
}

func parseOGMChapterTime(raw string) (int64, error) {
	var hours, minutes, seconds, millis int64
	parts := strings.Split(raw, ":")
	if len(parts) != 3 { return 0, fmt.Errorf("expected HH:MM:SS.mmm") }
	n, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || n < 0 || n > 1000 { return 0, fmt.Errorf("invalid chapter hour") }
	hours = n
	n, err = strconv.ParseInt(parts[1], 10, 64)
	if err != nil || n < 0 || n >= 60 { return 0, fmt.Errorf("invalid chapter minute") }
	minutes = n
	secs := strings.Split(parts[2], ".")
	if len(secs) != 2 || len(secs[1]) != 3 { return 0, fmt.Errorf("expected millisecond precision") }
	n, err = strconv.ParseInt(secs[0], 10, 64)
	if err != nil || n < 0 || n >= 60 { return 0, fmt.Errorf("invalid chapter second") }
	seconds = n
	millis, err = strconv.ParseInt(secs[1], 10, 64)
	if err != nil || millis < 0 || millis > 999 { return 0, fmt.Errorf("invalid chapter millisecond") }
	return ((hours*60+minutes)*60+seconds)*1000+millis, nil
}

// Returns matched=false for regular FFMETADATA1 / MKV chapter sources.
// OGM-style CHAPTER01=... / CHAPTER01NAME=... is strictly parsed so malformed
// input is rejected instead of generating incorrect or reordered chapter marks.
func parseOGMChapters(path string) (chapters []ogmChapter, matched bool, err error) {
	info, err := os.Stat(path)
	if err != nil { return nil, false, err }
	if info.Size() > 1024*1024 || info.IsDir() { return nil, false, nil }
	data, err := os.ReadFile(path)
	if err != nil { return nil, false, err }
	data = bytes.TrimPrefix(data, []byte{0xef, 0xbb, 0xbf})
	lines := bufio.NewScanner(bytes.NewReader(data))
	starts := map[int]int64{}
	names := map[int]string{}
	seen := false
	lineNo := 0
	for lines.Scan() {
		lineNo++
		line := strings.TrimSpace(lines.Text())
		if line == "" { continue }
		if !seen {
			if !strings.HasPrefix(strings.ToUpper(line), "CHAPTER") { return nil, false, nil }
			seen = true
		}
		match := ogmChapterLine.FindStringSubmatch(line)
		if match == nil { return nil, true, fmt.Errorf("OGM chapter line %d is invalid", lineNo) }
		id, _ := strconv.Atoi(match[1])
		if id < 1 { return nil, true, fmt.Errorf("chapter indexes start at 1") }
		if match[2] != "" {
			if _, exists := names[id]; exists { return nil, true, fmt.Errorf("duplicate chapter name %d", id) }
			names[id] = strings.TrimSpace(match[3])
		} else {
			if _, exists := starts[id]; exists { return nil, true, fmt.Errorf("duplicate chapter %d", id) }
			timestamp, e := parseOGMChapterTime(strings.TrimSpace(match[3]))
			if e != nil { return nil, true, fmt.Errorf("chapter %d: %w", id, e) }
			starts[id] = timestamp
		}
	}
	if err := lines.Err(); err != nil { return nil, matched, err }
	if !seen { return nil, false, nil }
	if len(starts) == 0 { return nil, true, fmt.Errorf("OGM chapter text contains no chapter timestamps") }
	for i := 1; i <= len(starts); i++ {
		start, ok := starts[i]
		if !ok { return nil, true, fmt.Errorf("missing chapter %d", i) }
		if i > 1 && start <= chapters[i-2].Start {
			return nil, true, fmt.Errorf("chapter %d must start after the previous chapter", i)
		}
		title := names[i]
		if title == "" { title = fmt.Sprintf("Chapter %d", i) }
		chapters = append(chapters, ogmChapter{Start: start, Title: title})
	}
	for id := range names {
		if _, ok := starts[id]; !ok { return nil, true, fmt.Errorf("chapter name %d has no timestamp", id) }
	}
	return chapters, true, nil
}

func escapeMergerFFMetadata(value string) string {
	r := strings.NewReplacer("\\", "\\\\", "=", "\\=", ";", "\\;", "#", "\\#", "\n", " ", "\r", " ")
	return r.Replace(value)
}

func ogmChaptersFFMetadata(chapters []ogmChapter, endMs int64) []byte {
	var b strings.Builder
	b.WriteString(";FFMETADATA1\n")
	for i, chapter := range chapters {
		end := chapter.Start + 1000 // fallback when input duration cannot be determined
		if i+1 < len(chapters) {
			end = chapters[i+1].Start
		} else if endMs > chapter.Start {
			end = endMs
		}
		fmt.Fprintf(&b, "[CHAPTER]\nTIMEBASE=1/1000\nSTART=%d\nEND=%d\ntitle=%s\n",
			chapter.Start, end, escapeMergerFFMetadata(chapter.Title))
	}
	return []byte(b.String())
}
