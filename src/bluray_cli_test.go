//go:build windows || linux

package main

import (
	"strings"
	"testing"
)

func TestBluRayCLIInventoryShowsPlaylistPIDAndLanguages(t *testing.T) {
	list := blurayCLIInventory(blurayPlaylist{Number: 800, DurationTicks: 45000*10822, Chapters: []blurayChapter{{Number:1}}},
		[]blurayStream{
			{Index:0,Kind:"video",Codec:"h264",PID:0x1011},
			{Index:1,Kind:"audio",Codec:"dts",Profile:"DTS-HD MA",PID:0x1100,Language:"eng",Channels:6},
			{Index:9,Kind:"subtitle",Codec:"hdmv_pgs_subtitle",PID:0x1204,Language:"nor"},
		})
	for _, want := range []string{"playlist 00800", "0x1011", "0x1100", "0x1204", "eng", "nor", "hdmv_pgs_subtitle", "DTS"} {
		if !strings.Contains(list,want) {t.Fatalf("Blu-ray CLI inventory missing %q: %s",want,list)}
	}
}

func TestBluRayCLIExampleCommandsAndParsing(t *testing.T) {
	examples:=blurayCLIExamples()
	if len(examples)!=4 {t.Fatalf("expected four discoverable Blu-ray presets: %v",examples)}
	expected:=[]string{"bluray scan", "bluray tracks --playlist 00800", "bluray remux --playlist 00800", "bluray demux --playlist 00800"}
	for i,line:=range examples {
		if !strings.Contains(line,expected[i]) {t.Fatalf("preset %d does not match its action: %s",i,line)}
		parts,err:=splitCLICommand(line)
		if err!=nil {t.Fatal(err)}
		if len(parts)<3 || parts[0]!="bluray" {t.Fatalf("command not executable via in-app CLI: %v",parts)}
	}
	raw:=[]string{"SOURCE","--playlist","00800","--output","OUTPUT_DIR","--video","both","--streams","0,1,9","--no-chapters"}
	args:=cliFlagsFirst(raw)
	expectedArgs:=[]string{"--playlist","00800","--output","OUTPUT_DIR","--video","both","--streams","0,1,9","--no-chapters","--","SOURCE"}
	if strings.Join(args,"|")!=strings.Join(expectedArgs,"|") { t.Fatalf("Blu-ray CLI flags reordered incorrectly: got %v expected %v",args,expectedArgs)}
	if err:=validateBlurayVideoMode("elementary");err!=nil {t.Fatal(err)}
	if err:=validateBlurayVideoMode("m2ts");err!=nil {t.Fatal(err)}
	if err:=validateBlurayVideoMode("both");err!=nil {t.Fatal(err)}
	for _,bad:=range []string{"vob","mpeg2","","unknown"} {
		if err:=validateBlurayVideoMode(bad);err==nil {t.Fatalf("DVD/unknown mode %q accepted by Blu-ray CLI",bad)}
	}
}

func TestBluRayCLIHelpDocumentsDirectExtractionAndM2TS(t *testing.T) {
	help:=blurayCLIUsage()
	for _,required:=range []string{"bluray scan","bluray tracks","bluray remux","bluray demux", "--playlist", "MOVIE.m2ts", "elementary|m2ts|both", "intermediate MKV", "no chapters"} {
		if !strings.Contains(help,required){t.Fatalf("Blu-ray CLI help missing %q: %s",required,help)}
	}
}
