//go:build windows || linux

package main

import (
	"reflect"
	"strings"
	"testing"
)

func TestNativeBlurayLanguagesMatchByPIDNotTrackOrder(t *testing.T) {
	raw := []byte("MUKSMATT_BD_LANG_1\n" +
		"S\ta\t4352\tENG\n" + // 0x1100
		"S\ta\t4353\tfra\n" + // 0x1101
		"S\ts\t4608\tnor\n" + // 0x1200
		"S\ta\t4352\teng\n")  // duplicate same clip language
	langs, err := parseNativeBlurayLanguages(raw)
	if err != nil { t.Fatal(err) }
	streams, err := parseBlurayProbe([]byte(`{"streams":[
		{"index":0,"id":"0x1011","codec_type":"video","codec_name":"h264"},
		{"index":1,"id":"0x1101","codec_type":"audio","codec_name":"dts","channels":6},
		{"index":2,"id":"0x1100","codec_type":"audio","codec_name":"pcm_bluray","channels":2},
		{"index":3,"id":"0x1200","codec_type":"subtitle","codec_name":"hdmv_pgs_subtitle"},
		{"index":4,"codec_type":"subtitle","codec_name":"hdmv_pgs_subtitle"}
	]}`))
	if err != nil { t.Fatal(err) }
	addBlurayLanguages(streams, langs)
	got := []string{streams[0].Language, streams[1].Language, streams[2].Language, streams[3].Language, streams[4].Language}
	if !reflect.DeepEqual(got, []string{"", "fra", "eng", "nor", ""}) {
		t.Fatalf("languages assigned to wrong stream: %v", got)
	}
	opts := blurayMappedStreamMetadataArgs(streams)
	joined := strings.Join(opts, " ")
	for _, required := range []string{"-metadata:s:a:0 language=fra", "-metadata:s:a:1 language=eng", "-metadata:s:s:0 language=nor", "title=PGS subtitles"} {
		if !strings.Contains(joined, required) {t.Fatalf("missing %q in %s", required, joined)}
	}
	if strings.Contains(joined, "-metadata:s:s:1 language=") {
		t.Fatalf("unknown subtitle language must not be fabricated: %s", joined)
	}
}

func TestNativeBlurayLanguageConflictsAreNotAssigned(t *testing.T) {
	langs, err := parseNativeBlurayLanguages([]byte("MUKSMATT_BD_LANG_1\nS\ta\t4352\teng\nS\ta\t4352\tfre\nS\ta\t4352\teng\nS\ts\t4608\tund\n"))
	if err != nil {t.Fatal(err)}
	if _, ok := langs[blurayLanguageKey{kind: "audio", pid: 4352}]; ok {t.Fatal("conflicting playlist languages must remain unknown")}
	if len(langs) != 0 {t.Fatalf("unexpected languages: %#v", langs)}
	for _, raw := range []string{
		"wrong\nS\ta\t4352\teng\n",
		"MUKSMATT_BD_LANG_1\nS\ta\t8192\teng\n",
		"MUKSMATT_BD_LANG_1\nS\ts\t4352\t123\n",
		"MUKSMATT_BD_LANG_1\nS\tx\t4352\teng\n",
	} {
		if _, err := parseNativeBlurayLanguages([]byte(raw)); err == nil {
			t.Fatalf("accepted invalid metadata: %q", raw)
		}
	}
}

func TestBlurayMetadataWithSelectedStreamsOnly(t *testing.T) {
	source := testBlurayDiscRoot(t)
	all := []blurayStream{
		{Index:0,Kind:"video",Codec:"h264"},
		{Index:1,Kind:"audio",Codec:"dts",Language:"eng",Channels:6},
		{Index:2,Kind:"audio",Codec:"dts",Language:"fra",Channels:6},
		{Index:3,Kind:"subtitle",Codec:"hdmv_pgs_subtitle",Language:"nor"},
	}
	selected, err := selectedBlurayStreams(all, []int{2,3})
	if err != nil {t.Fatal(err)}
	args, err := blurayRemuxArgs(source, blurayPlaylist{Number:800}, selected, "", "partial.mkv")
	if err != nil {t.Fatal(err)}
	command := strings.Join(args, " ")
	for _, required := range []string{"-map 0:2", "-map 0:3", "-metadata:s:a:0 language=fra", "-metadata:s:s:0 language=nor"} {
		if !strings.Contains(command, required) {t.Fatalf("missing %q in %s", required, command)}
	}
	if strings.Contains(command, "language=eng") || strings.Contains(command, "-map 0:1") {
		t.Fatalf("unselected language or stream leaked: %s", command)
	}
}

func TestAdvancedMergerBlurayTrackLanguagesRespectOutputOrder(t *testing.T) {
	streams := []mergerStream{
		{Path:"ordinary.mkv",Track:trackOption{Kind:"audio",Index:2,Language:"ita"}},
		{Path:"blu-ray",Track:trackOption{Kind:"audio",Index:1,Language:"eng",Title:"DTS (6 channels)"},BluRay:true,BluRayPlaylist:800},
		{Path:"blu-ray",Track:trackOption{Kind:"subtitle",Index:8,Language:"nor",Title:"PGS subtitles"},BluRay:true,BluRayPlaylist:800},
	}
	flags := strings.Join(mergerBlurayMetadataArgs(streams), " ")
	for _, want := range []string{"-metadata:s:a:1 language=eng", "-metadata:s:s:0 language=nor"} {
		if !strings.Contains(flags,want) {t.Fatalf("missing %q in %s",want,flags)}
	}
	if strings.Contains(flags, "language=ita") {t.Fatalf("ordinary source should retain original tags: %s",flags)}
}

func TestVerifyBlurayRemuxStreamsMetadata(t *testing.T) {
	selected := []blurayStream{
		{Index: 0, Kind: "video", Codec: "h264"},
		{Index: 1, Kind: "audio", Codec: "dts", Profile: "DTS-HD MA", Language: "eng", Channels: 6},
		{Index: 2, Kind: "audio", Codec: "pcm_bluray", Language: "fra", Channels: 2},
		{Index: 3, Kind: "subtitle", Codec: "hdmv_pgs_subtitle", Language: "nor"},
		{Index: 4, Kind: "subtitle", Codec: "hdmv_pgs_subtitle"}, // unknown language remains unknown
	}
	output := []blurayStream{
		{Index: 0, Kind: "video", Codec: "h264"},
		{Index: 1, Kind: "audio", Codec: "dts", Language: "eng", Title: "DTS-HD MA (6 channels)"},
		{Index: 2, Kind: "audio", Codec: "flac", Language: "fra", Title: "LPCM (lossless FLAC in MKV) (2 channels)"},
		{Index: 3, Kind: "subtitle", Codec: "hdmv_pgs_subtitle", Language: "nor", Title: "PGS subtitles"},
		{Index: 4, Kind: "subtitle", Codec: "hdmv_pgs_subtitle", Language: "und", Title: "PGS subtitles"},
	}
	if err := verifyBlurayRemuxStreams(output, selected); err != nil {
		t.Fatalf("correctly tagged MKV must pass: %v", err)
	}
	cases := []struct {
		name string
		index int
		change func(*blurayStream)
		want string
	}{
		{"missing audio language", 1, func(s *blurayStream) { s.Language = "" }, "language mismatch"},
		{"incorrect subtitle language", 3, func(s *blurayStream) { s.Language = "eng" }, "language mismatch"},
		{"missing descriptive audio title", 1, func(s *blurayStream) { s.Title = "" }, "title mismatch"},
		{"incorrect subtitle title", 3, func(s *blurayStream) { s.Title = "Other" }, "title mismatch"},
		{"LPCM not converted to FLAC", 2, func(s *blurayStream) { s.Codec = "pcm_bluray" }, "mismatch"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mutated := append([]blurayStream(nil), output...)
			tc.change(&mutated[tc.index])
			if err := verifyBlurayRemuxStreams(mutated, selected); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("expected %q, got %v", tc.want, err)
			}
		})
	}
	if err := verifyBlurayRemuxStreams(output[:len(output)-1], selected); err == nil || !strings.Contains(err.Error(), "stream count mismatch") {
		t.Fatalf("expected stream-count failure, got %v", err)
	}
}

/*
TestBD50Playlist00800LanguagePIDMapping records the 18 language/PID pairs
observed on a real, decrypted BD50. It exercises mapping without requiring
the physical disc or assigning languages by index.
*/
func TestBD50Playlist00800LanguagePIDMapping(t *testing.T) {
	native := []byte("MUKSMATT_BD_LANG_1\n" +
		"S\ta\t4352\teng\n" +
		"S\ta\t4353\tspa\n" +
		"S\ta\t4354\tces\n" +
		"S\ta\t4355\tpol\n" +
		"S\ts\t4608\teng\n" +
		"S\ts\t4609\tspa\n" +
		"S\ts\t4610\tdan\n" +
		"S\ts\t4611\tfin\n" +
		"S\ts\t4612\tnor\n" +
		"S\ts\t4613\tswe\n" +
		"S\ts\t4614\tces\n" +
		"S\ts\t4615\tpol\n" +
		"S\ts\t4616\tzho\n" +
		"S\ts\t4617\tzho\n" +
		"S\ts\t4618\tkor\n" +
		"S\ts\t4619\tspa\n" +
		"S\ts\t4620\tces\n" +
		"S\ts\t4621\tpol\n")
	languages, err := parseNativeBlurayLanguages(native)
	if err != nil {
		t.Fatal(err)
	}
	if len(languages) != 18 {
		t.Fatalf("got %d language mappings, want 18", len(languages))
	}

	probe := []byte(`{"streams":[
		{"index":0,"id":"0x1011","codec_type":"video","codec_name":"h264"},
		{"index":1,"id":"0x1100","codec_type":"audio","codec_name":"dts","profile":"DTS-HD MA"},
		{"index":2,"id":"0x1101","codec_type":"audio","codec_name":"dts","profile":"DTS"},
		{"index":3,"id":"0x1102","codec_type":"audio","codec_name":"dts","profile":"DTS"},
		{"index":4,"id":"0x1103","codec_type":"audio","codec_name":"dts","profile":"DTS"},
		{"index":5,"id":"0x1200","codec_type":"subtitle","codec_name":"hdmv_pgs_subtitle"},
		{"index":6,"id":"0x1201","codec_type":"subtitle","codec_name":"hdmv_pgs_subtitle"},
		{"index":7,"id":"0x1202","codec_type":"subtitle","codec_name":"hdmv_pgs_subtitle"},
		{"index":8,"id":"0x1203","codec_type":"subtitle","codec_name":"hdmv_pgs_subtitle"},
		{"index":9,"id":"0x1204","codec_type":"subtitle","codec_name":"hdmv_pgs_subtitle"},
		{"index":10,"id":"0x1205","codec_type":"subtitle","codec_name":"hdmv_pgs_subtitle"},
		{"index":11,"id":"0x1206","codec_type":"subtitle","codec_name":"hdmv_pgs_subtitle"},
		{"index":12,"id":"0x1207","codec_type":"subtitle","codec_name":"hdmv_pgs_subtitle"},
		{"index":13,"id":"0x1208","codec_type":"subtitle","codec_name":"hdmv_pgs_subtitle"},
		{"index":14,"id":"0x1209","codec_type":"subtitle","codec_name":"hdmv_pgs_subtitle"},
		{"index":15,"id":"0x120a","codec_type":"subtitle","codec_name":"hdmv_pgs_subtitle"},
		{"index":16,"id":"0x120b","codec_type":"subtitle","codec_name":"hdmv_pgs_subtitle"},
		{"index":17,"id":"0x120c","codec_type":"subtitle","codec_name":"hdmv_pgs_subtitle"},
		{"index":18,"id":"0x120d","codec_type":"subtitle","codec_name":"hdmv_pgs_subtitle"}
	]}`)
	streams, err := parseBlurayProbe(probe)
	if err != nil {
		t.Fatal(err)
	}
	addBlurayLanguages(streams, languages)

	want := []string{
		"", "eng", "spa", "ces", "pol",
		"eng", "spa", "dan", "fin", "nor", "swe", "ces", "pol",
		"zho", "zho", "kor", "spa", "ces", "pol",
	}
	if len(streams) != len(want) {
		t.Fatalf("got %d streams, want %d", len(streams), len(want))
	}
	for i, stream := range streams {
		if stream.Language != want[i] {
			t.Errorf("stream %d PID %#x: got language %q, want %q", i, stream.PID, stream.Language, want[i])
		}
	}
	metadata := strings.Join(blurayMappedStreamMetadataArgs(streams), " ")
	for _, item := range []string{
		"-metadata:s:a:0 language=eng",
		"-metadata:s:a:1 language=spa",
		"-metadata:s:a:2 language=ces",
		"-metadata:s:a:3 language=pol",
		"-metadata:s:s:4 language=nor",
		"-metadata:s:s:8 language=zho",
		"-metadata:s:s:9 language=zho",
		"-metadata:s:s:13 language=pol",
	} {
		if !strings.Contains(metadata, item) {
			t.Errorf("missing metadata mapping %q", item)
		}
	}
}
