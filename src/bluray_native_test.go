//go:build windows || linux

package main

import (
	"strings"
	"testing"
)

func TestNativeBlurayPlaylistMetadataAndChapters(t *testing.T) {
	payload := "MUKSMATT_BD_NAV_1\n" +
		"P\t800\t5400000\t2\n" +
		"C\t800\t0\t3600000\n" +
		"C\t800\t3600000\t5400000\n" +
		"P\t2\t900000\t1\n" +
		"C\t2\t0\t900000\n"
	playlists, err := parseNativeBlurayNavigation([]byte(payload))
	if err != nil { t.Fatal(err) }
	if len(playlists) != 2 || playlists[0].Number != 2 || playlists[1].Number != 800 {
		t.Fatalf("wrong native playlist order: %+v", playlists)
	}
	best, err := selectBlurayPlaylist(playlists, "")
	if err != nil || best.Number != 800 || best.DurationTicks != 2700000 {
		t.Fatalf("wrong longest native playlist: %+v, err %v", best, err)
	}
	if len(best.Chapters) != 2 ||
		best.Chapters[1].StartTicks != 1800000 ||
		best.Chapters[1].EndTicks != 2700000 {
		t.Fatalf("wrong native 90kHz-to-45kHz chapter conversion: %+v", best.Chapters)
	}
}

func TestNativeBlurayNavigationRejectsUnsafeRecords(t *testing.T) {
	failures := []string{
		"",
		"UNSUPPORTED\nP\t800\t900000\t1\n",
		"MUKSMATT_BD_NAV_1\nP\t800\t900000\t1\nP\t800\t900000\t1\n",
		"MUKSMATT_BD_NAV_1\nP\t800\t900000\t0\n",
		"MUKSMATT_BD_NAV_1\nP\t800\t900000\t1\nC\t999\t0\t90000\n",
		"MUKSMATT_BD_NAV_1\nP\t800\t900000\t1\nC\t800\t-1\t90000\n",
		"MUKSMATT_BD_NAV_1\nP\t800\t900000\t1\nC\t800\t0\t900001\n",
		"MUKSMATT_BD_NAV_1\nP\t800\t900000\t1\nC\t800\t100000\t100000\n",
		"MUKSMATT_BD_NAV_1\nP\t800\t900000\t1\nC\t800\t0\t540000\nC\t800\t450000\t900000\n",
		"MUKSMATT_BD_NAV_1\nP\t800\t900000\t1\ninvalid\n",
	}
	for i, payload := range failures {
		if _, err := parseNativeBlurayNavigation([]byte(payload)); err == nil {
			t.Fatalf("case %d accepted invalid native navigation:\n%s", i, payload)
		}
	}
	if _, err := parseNativeBlurayNavigation([]byte(strings.Repeat("x", 8<<20+1))); err == nil {
		t.Fatal("oversized native navigation accepted")
	}
}

func TestBlurayISOProtocolAndSourceSafety(t *testing.T) {
	iso := bluraySource{Kind: bluraySourceISOCandidate, Input: "Movie.iso"}
	args, err := blurayInputArgs(iso, blurayPlaylist{Number: 800})
	if err != nil || !strings.HasSuffix(args[3], "Movie.iso") {
		t.Fatalf("ISO libbluray protocol: %v %v", args, err)
	}
	drive := bluraySource{Kind: bluraySourcePhysicalDrive, Input: "/dev/sr0"}
	args, err = blurayInputArgs(drive, blurayPlaylist{Number: 10})
	if err != nil || !strings.Contains(args[3], "/dev/sr0") {
		t.Fatalf("optical libbluray protocol: %v %v", args, err)
	}
	for _, unsupported := range []bluraySource{{}, {Kind: bluraySourceDirectory, Input: "not-a-root"}} {
		if _, err := blurayProtocolURL(unsupported); err == nil {
			t.Fatalf("accepted unsupported source: %#v", unsupported)
		}
	}
}
