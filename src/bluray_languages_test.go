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
