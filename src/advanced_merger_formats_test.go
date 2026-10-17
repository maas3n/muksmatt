//go:build windows || linux

package main

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestAdvancedOGMChaptersAutodetection(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ChaptersNew.txt")
	content := "CHAPTER01=00:00:00.000\nCHAPTER01NAME=hei1\nCHAPTER02=00:07:23.526\nCHAPTER02NAME=hei2\nCHAPTER03=02:54:12.442\nCHAPTER03NAME=hei3\n"
	if err := os.WriteFile(path, []byte(content), 0600); err != nil { t.Fatal(err) }
	chapters, matched, err := parseOGMChapters(path)
	if err != nil || !matched || len(chapters) != 3 { t.Fatalf("OGM parse: %v %t %v", chapters, matched, err) }
	if chapters[1].Start != 443526 || chapters[2].Start != 10452442 {
		t.Fatalf("chapter timestamps changed: %v", chapters)
	}
	if err := validateMergerChapters(context.Background(), "missing-ffprobe", path); err != nil {
		t.Fatalf("OGM must validate without external probing: %v", err)
	}
	out := string(ogmChaptersFFMetadata(chapters, 10822144))
	if !strings.Contains(out, "START=10452442\nEND=10822144\ntitle=hei3") || strings.Count(out, "[CHAPTER]") != 3 {
		t.Fatalf("converted chapter metadata wrong: %s", out)
	}
	for _, invalid := range []string{
		"CHAPTER01=00:01:00.000\nCHAPTER02=00:00:00.000\n",
		"CHAPTER02=00:00:00.000\n",
		"CHAPTER01=00:00:00.000\nCHAPTER01=00:01:00.000\n",
		"CHAPTER01=00:00:00.000\nCHAPTER03NAME=unknown\n",
	} {
		if err := os.WriteFile(path, []byte(invalid), 0600); err != nil { t.Fatal(err) }
		_, detected, err := parseOGMChapters(path)
		if !detected || err == nil { t.Fatalf("invalid OGM accepted %q: %v", invalid, err) }
	}
}

func TestAdvancedMuxMetadataEdits(t *testing.T) {
	tracks := []mergerStream{
		{Path:"video.h264", Track:trackOption{Index:0,Kind:"video",Codec:"h264"}},
		{Path:"ac3.ac3",Track:trackOption{Index:0,Kind:"audio",Codec:"ac3"}},
		{Path:"subrip.srt",Track:trackOption{Index:0,Kind:"subtitle",Codec:"subrip"}},
	}
	if err := setAdvancedStreamMetadata(&tracks[1], "ENG", "Dolby Digital stereo"); err != nil { t.Fatal(err) }
	if err := setAdvancedStreamMetadata(&tracks[2], "NOR", "Norwegian (OCR)"); err != nil { t.Fatal(err) }
	if tracks[1].Track.Language != "eng" || tracks[2].Track.Language != "nor" { t.Fatal("languages not normalized") }
	args, err := mergerArgs(tracks, "", "out.mkv")
	if err != nil { t.Fatal(err) }
	for _, want := range []string{
		"-metadata:s:a:0 language=eng",
		"-metadata:s:a:0 title=Dolby Digital stereo",
		"-metadata:s:s:0 language=nor",
		"-metadata:s:s:0 title=Norwegian (OCR)",
	} {
		if !strings.Contains(strings.Join(args, " "), want) { t.Fatalf("missing %s: %v", want, args) }
	}
	before := tracks[1]
	if err := setAdvancedStreamMetadata(&tracks[1], "invalid", "Title"); err == nil || !reflect.DeepEqual(before,tracks[1]) {
		t.Fatal("invalid edit mutated existing stream")
	}
}

func TestAdvancedM2TSAllStreamsExceptChapters(t *testing.T) {
	rows := []mergerStream{
		{Path:"source.m2ts",Track:trackOption{Index:0,Kind:"video",Codec:"h264"}},
		{Path:"source.m2ts",Track:trackOption{Index:1,Kind:"audio",Codec:"dts"}},
		{Path:"source.m2ts",Track:trackOption{Index:2,Kind:"subtitle",Codec:"hdmv_pgs_subtitle"}},
		{Path:"source.m2ts",Track:trackOption{Index:-1,Kind:"chapters",Title:"20 chapters"}},
	}
	if err := validateAdvancedM2TSStreams(rows); err != nil { t.Fatal(err) }
	args, err := mergerArgsForFormat(rows, "", "output.m2ts", "mpegts")
	if err != nil { t.Fatal(err) }
	all := strings.Join(args, " ")
	for _, want := range []string{"-map 0:0", "-map 0:1", "-map 0:2", "-map_chapters -1", "-mpegts_m2ts_mode 1 -f mpegts output.m2ts"} {
		if !strings.Contains(all, want) { t.Fatalf("missing %s in %s", want,all) }
	}
	for _, forbidden := range []string{" -c:a:0 flac", " -f matroska", " -map 0:-1"} {
		if strings.Contains(all, forbidden) { t.Fatalf("M2TS should not use %s: %s", forbidden,all) }
	}
	rows = append(rows, mergerStream{Path:"Norwegian.srt",Track:trackOption{Index:0,Kind:"subtitle",Codec:"subrip"}})
	if err := validateAdvancedM2TSStreams(rows); err == nil || !strings.Contains(err.Error(), "PGS") {
		t.Fatalf("SRT must fail safely: %v", err)
	}
}

func TestAdvancedM2TSIntegrationVideoAudioNoChapters(t *testing.T) {
	ffmpeg, err := exec.LookPath("ffmpeg")
	if err != nil { t.Skip("FFmpeg unavailable") }
	ffprobe, err := exec.LookPath("ffprobe")
	if err != nil { t.Skip("FFprobe unavailable") }
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer cancel()
	dir := t.TempDir()
	source := filepath.Join(dir, "source.mkv")
	if _, err := runMergerCommand(ctx, ffmpeg, "-v", "error", "-f", "lavfi", "-i",
		"color=size=64x64:rate=25:duration=2", "-f", "lavfi", "-i", "sine=duration=2",
		"-map", "0:v", "-map", "1:a", "-c:v", "mpeg2video", "-c:a", "ac3",
		source); err != nil { t.Fatal(err) }
	rows, err := probeMergerFile(ctx, ffprobe, source, "all")
	if err != nil { t.Fatal(err) }
	out := filepath.Join(dir, "test.m2ts")
	if err = muxMerger(ctx, toolPaths{ffmpeg:ffmpeg,ffprobe:ffprobe},rows,"",out); err != nil { t.Fatal(err) }
	data, err := os.ReadFile(out)
	if err != nil { t.Fatal(err) }
	if len(data)<192 || len(data)%192 !=0 { t.Fatalf("expected 192-byte M2TS packets: size=%d",len(data)) }
	for offset:=0;offset<len(data);offset+=192 {
		if data[offset+4]!=0x47 { t.Fatalf("invalid M2TS sync at %d", offset) }
	}
	probe, err := runMergerCommand(ctx, ffprobe,"-v","error","-show_streams","-show_chapters","-of","json",out)
	if err != nil { t.Fatal(err) }
	var p struct {
		Streams []struct {CodecType string `json:"codec_type"`; CodecName string `json:"codec_name"`} `json:"streams"`
		Chapters []json.RawMessage `json:"chapters"`
	}
	if err := json.Unmarshal(probe,&p); err != nil { t.Fatal(err) }
	if len(p.Streams)!=2 || p.Streams[0].CodecType!="video" || p.Streams[1].CodecName!="ac3" || len(p.Chapters)!=0 {
		t.Fatalf("M2TS stream inventory mismatch: %s",probe)
	}
}

func TestAdvancedMuxOGMChaptersAndImportedTrackLanguageIntegration(t *testing.T) {
	ffmpeg, err := exec.LookPath("ffmpeg")
	if err != nil { t.Skip("FFmpeg unavailable") }
	ffprobe, err := exec.LookPath("ffprobe")
	if err != nil { t.Skip("FFprobe unavailable") }
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer cancel()
	dir := t.TempDir()
	movie := filepath.Join(dir, "movie.mkv")
	srt := filepath.Join(dir, "Norwegian.srt")
	ogm := filepath.Join(dir, "ChaptersNew.txt")
	if err := os.WriteFile(srt, []byte("1\n00:00:00,000 --> 00:00:01,700\nHei!\n"), 0600); err != nil { t.Fatal(err) }
	if err := os.WriteFile(ogm, []byte("CHAPTER01=00:00:00.000\nCHAPTER01NAME=hei1\nCHAPTER02=00:00:00.500\nCHAPTER02NAME=hei2\n"), 0600); err != nil { t.Fatal(err) }
	if _, err := runMergerCommand(ctx, ffmpeg,
		"-v", "error", "-f", "lavfi", "-i", "color=size=64x64:rate=25:duration=2",
		"-f", "lavfi", "-i", "sine=duration=2",
		"-map", "0:v", "-map", "1:a", "-c:v", "mpeg2video", "-c:a", "ac3",
		movie); err != nil { t.Fatal(err) }
	movieTracks, err := probeMergerFile(ctx, ffprobe, movie, "all")
	if err != nil { t.Fatal(err) }
	sub, err := probeMergerFile(ctx, ffprobe, srt, "subtitle")
	if err != nil { t.Fatal(err) }
	merged := append(movieTracks, sub...)
	for i := range merged {
		if merged[i].Track.Kind == "audio" {
			if err := setAdvancedStreamMetadata(&merged[i], "eng", "English Dolby Digital"); err != nil { t.Fatal(err) }
		}
		if merged[i].Track.Kind == "subtitle" {
			if err := setAdvancedStreamMetadata(&merged[i], "nor", "Norwegian OCR"); err != nil { t.Fatal(err) }
		}
	}
	out := filepath.Join(dir, "done.mkv")
	if err := muxMerger(ctx, toolPaths{ffmpeg:ffmpeg,ffprobe:ffprobe},merged,ogm,out); err != nil { t.Fatal(err) }
	payload, err := runMergerCommand(ctx, ffprobe,"-v","error","-show_streams","-show_chapters","-of","json",out)
	if err != nil { t.Fatal(err) }
	var probed struct {
		Streams []struct { CodecType string `json:"codec_type"`; Tags map[string]string `json:"tags"` } `json:"streams"`
		Chapters []struct {Tags map[string]string `json:"tags"`} `json:"chapters"`
	}
	if err := json.Unmarshal(payload,&probed); err != nil { t.Fatal(err) }
	if len(probed.Streams)!=3 || len(probed.Chapters)!=2 {
		t.Fatalf("output missing chapter rows or tracks: %s",payload)
	}
	if probed.Chapters[0].Tags["title"]!="hei1" || probed.Chapters[1].Tags["title"]!="hei2" {
		t.Fatalf("chapter names missing: %s",payload)
	}
	for _, row := range probed.Streams {
		switch row.CodecType {
		case "audio":
			if row.Tags["language"]!="eng" || row.Tags["title"]!="English Dolby Digital" { t.Fatalf("audio metadata missing: %s",payload) }
		case "subtitle":
			if row.Tags["language"]!="nor" || row.Tags["title"]!="Norwegian OCR" { t.Fatalf("SRT metadata missing: %s",payload) }
		}
	}
}
