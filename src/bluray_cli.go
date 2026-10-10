//go:build windows || linux

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
)

// Explicit Blu-ray subcommands protect the existing DVD scan/remux syntax.
// ISO/physical-device playlist enumeration requires the companion built with
// native libbluray; accessible BDMV folders retain their Go-only MPLS parser.
func cliBluray(ctx context.Context, args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: muksmatt-cli bluray scan ROOT | bluray remux --output MOVIE.mkv [--playlist 00800] [--streams 0,1] [--no-chapters] ROOT")
		os.Exit(2)
	}
	switch args[0] {
	case "scan":
		fs := flag.NewFlagSet("bluray scan", flag.ExitOnError)
		_ = fs.Parse(cliFlagsFirst(args[1:]))
		if fs.NArg() != 1 {
			fmt.Fprintln(os.Stderr, "Usage: muksmatt-cli bluray scan SOURCE")
			os.Exit(2)
		}
		source, err := resolveBluraySource(fs.Arg(0))
		fatalIf(err)
		playlists, err := discoverBlurayForSource(ctx, source)
		fatalIf(err)
		best, err := selectBlurayPlaylist(playlists, "")
		fatalIf(err)
		fmt.Printf("%-10s %-14s %-10s %s\n", "PLAYLIST", "DURATION", "CHAPTERS", "DEFAULT")
		for _, p := range playlists {
			selected := ""
			if p.Number == best.Number {
				selected = "longest"
			}
			fmt.Printf("%05d      %-14s %-10d %s\n", p.Number, p.Duration(), len(p.Chapters), selected)
		}
	case "remux":
		fs := flag.NewFlagSet("bluray remux", flag.ExitOnError)
		id := fs.String("playlist", "", "MPLS playlist ID; default chooses longest")
		output := fs.String("output", "", "destination .mkv file (required)")
		streams := fs.String("streams", "", "comma-separated absolute input stream indexes (default: all A/V/subtitle)")
		noChapters := fs.Bool("no-chapters", false, "omit Blu-ray MPLS chapters")
		_ = fs.Parse(cliFlagsFirst(args[1:]))
		if fs.NArg() != 1 || strings.TrimSpace(*output) == "" {
			fmt.Fprintln(os.Stderr, "Usage: muksmatt-cli bluray remux --output MOVIE.mkv [--playlist 00800] [--streams 0,1] [--no-chapters] SOURCE")
			os.Exit(2)
		}
		source, err := resolveBluraySource(fs.Arg(0))
		fatalIf(err)
		// ISO images and physical drives require native libbluray playlist discovery.
		// Folder sources continue using the bounded MPLS parser.
		playlists, err := discoverBlurayForSource(ctx, source)
		fatalIf(err)
		playlist, err := selectBlurayPlaylist(playlists, *id)
		fatalIf(err)
		selected, err := parseCLIStreams(*streams)
		fatalIf(err)
		tools, err := resolveBlurayTools(ctx)
		fatalIf(err)
		fmt.Fprintf(os.Stderr, "Blu-ray playlist %05d (%s), FFmpeg: %s\n", playlist.Number, playlist.Duration(), tools.ffmpeg)
		if !*noChapters {
			fmt.Fprintf(os.Stderr, "Preserving %d MPLS chapter marks.\n", len(playlist.Chapters))
		}
		fatalIf(remuxBlurayPlaylist(ctx, tools, source, playlist, *output, selected, !*noChapters))
		fmt.Println(*output)
	default:
		fmt.Fprintf(os.Stderr, "unknown Blu-ray subcommand %q\n", args[0])
		os.Exit(2)
	}
}
