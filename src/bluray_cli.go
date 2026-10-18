//go:build windows || linux

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Blu-ray CLI uses the same libbluray playlist reader, PID/language matching,
// MKV remux, M2TS mux and direct elementary demux as the GUI tabs.
// DVD CLI syntax and input flags are intentionally unchanged.
func cliBluray(ctx context.Context, args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, blurayCLIUsage())
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
		for _, playlist := range playlists {
			mark := ""
			if playlist.Number == best.Number { mark = "longest" }
			fmt.Printf("%05d      %-14s %-10d %s\n", playlist.Number, playlist.Duration(), len(playlist.Chapters), mark)
		}
		fmt.Fprintln(os.Stderr, "Select the intended playlist explicitly with --playlist (longest can be a looping/decoy playlist).")
	case "tracks", "metadata", "info":
		fs := flag.NewFlagSet("bluray tracks", flag.ExitOnError)
		id := fs.String("playlist", "", "explicit MPLS playlist ID (e.g. 00800)")
		_ = fs.Parse(cliFlagsFirst(args[1:]))
		if fs.NArg() != 1 || strings.TrimSpace(*id) == "" {
			fmt.Fprintln(os.Stderr, "Usage: muksmatt-cli bluray tracks --playlist 00800 SOURCE")
			os.Exit(2)
		}
		source, playlist, err := cliSelectedBlurayPlaylist(ctx, fs.Arg(0), *id)
		fatalIf(err)
		tools, err := resolveBlurayTools(ctx)
		fatalIf(err)
		streams, err := probeBlurayPlaylist(ctx, tools, source, playlist)
		fatalIf(err)
		fmt.Print(blurayCLIInventory(playlist, streams))
	case "remux":
		fs := flag.NewFlagSet("bluray remux", flag.ExitOnError)
		id := fs.String("playlist", "", "MPLS playlist ID; if omitted chooses longest (may be a decoy)")
		output := fs.String("output", "", "output .mkv or .m2ts file (required)")
		streams := fs.String("streams", "", "comma-separated absolute stream indexes (default: all)")
		noChapters := fs.Bool("no-chapters", false, "omit chapters (M2TS always excludes them)")
		_ = fs.Parse(cliFlagsFirst(args[1:]))
		if fs.NArg() != 1 || strings.TrimSpace(*output) == "" {
			fmt.Fprintln(os.Stderr, "Usage: muksmatt-cli bluray remux --playlist 00800 --output MOVIE.mkv|MOVIE.m2ts [--streams 0,1] [--no-chapters] SOURCE")
			os.Exit(2)
		}
		source, playlist, err := cliSelectedBlurayPlaylist(ctx, fs.Arg(0), *id)
		fatalIf(err)
		if *id == "" {
			fmt.Fprintln(os.Stderr, "Warning: longest Blu-ray playlist can be a looping decoy. Scan the disc and use --playlist for reliable selection.")
		}
		selected, err := parseCLIStreams(*streams)
		fatalIf(err)
		tools, err := resolveBlurayTools(ctx)
		fatalIf(err)
		fmt.Fprintf(os.Stderr, "Blu-ray playlist %05d (%s), FFmpeg: %s\n", playlist.Number, playlist.Duration(), tools.ffmpeg)
		switch strings.ToLower(filepath.Ext(*output)) {
		case ".mkv":
			if !*noChapters {
				fmt.Fprintf(os.Stderr, "Preserving %d MPLS chapter marks.\n", len(playlist.Chapters))
			}
			fatalIf(remuxBlurayPlaylist(ctx, tools, source, playlist, *output, selected, !*noChapters))
		case ".m2ts":
			fmt.Fprintln(os.Stderr, "Muxing selected video/audio/PGS into M2TS. Chapters are not supported and will be omitted.")
			fatalIf(cliBlurayMuxM2TS(ctx, tools, source, playlist, *output, selected))
		default:
			fatalIf(fmt.Errorf("Blu-ray remux --output must end in .mkv or .m2ts"))
		}
		fmt.Println(*output)
	case "demux":
		fs := flag.NewFlagSet("bluray demux", flag.ExitOnError)
		id := fs.String("playlist", "", "explicit MPLS playlist ID (e.g. 00800)")
		output := fs.String("output", "", "existing output directory (required)")
		streams := fs.String("streams", "", "comma-separated absolute indexes; defaults to all media streams")
		video := fs.String("video", blurayVideoElementary, "elementary, m2ts or both; audio/subtitles are exported individually")
		noChapters := fs.Bool("no-chapters", false, "omit Chapters.txt")
		_ = fs.Parse(cliFlagsFirst(args[1:]))
		if fs.NArg() != 1 || strings.TrimSpace(*id) == "" || strings.TrimSpace(*output) == "" {
			fmt.Fprintln(os.Stderr, "Usage: muksmatt-cli bluray demux --playlist 00800 --output DIR [--streams 0,1,9] [--video elementary|m2ts|both] [--no-chapters] SOURCE")
			os.Exit(2)
		}
		_, playlist, err := cliSelectedBlurayPlaylist(ctx, fs.Arg(0), *id)
		fatalIf(err)
		indexes, err := parseCLIStreams(*streams)
		fatalIf(err)
		fatalIf(validateBlurayVideoMode(*video))
		fmt.Fprintf(os.Stderr, "Directly demuxing playlist %05d; no temporary MKV will be created.\n", playlist.Number)
		final, err := guiBlurayDemuxWithVideoMode(ctx, fs.Arg(0), playlist.Number, *output, indexes, !*noChapters, *video)
		fatalIf(err)
		fmt.Println(final)
	default:
		fmt.Fprintf(os.Stderr, "unknown Blu-ray subcommand %q\n\n%s\n", args[0], blurayCLIUsage())
		os.Exit(2)
	}
}

func blurayCLIUsage() string {
	return `Blu-ray CLI (Windows and Linux):
  muksmatt-cli bluray scan SOURCE
  muksmatt-cli bluray tracks --playlist 00800 SOURCE
  muksmatt-cli bluray remux --playlist 00800 --output MOVIE.mkv SOURCE
  muksmatt-cli bluray remux --playlist 00800 --output MOVIE.m2ts SOURCE
  muksmatt-cli bluray demux --playlist 00800 --output EXISTING_DIR [--streams 0,1,9] [--video elementary|m2ts|both] SOURCE

Use scan first to select the correct MPLS playlist; the longest can be fake.
M2TS remux includes compatible video/audio/PGS streams but no chapters or SRT text subtitles.
DEMUX reads directly through libbluray and produces separate elementary tracks, without an intermediate MKV.
--video m2ts means video-only M2TS plus separate audio/subtitle files.
--video both exports both native video and video-only M2TS alongside audio/subtitle files.`
}

func blurayCLIExamples() []string {
	return []string{
		`muksmatt-cli bluray scan "SOURCE"`,
		`muksmatt-cli bluray tracks --playlist 00800 "SOURCE"`,
		`muksmatt-cli bluray remux --playlist 00800 --output "movie.mkv" "SOURCE"`,
		`muksmatt-cli bluray demux --playlist 00800 --output "OUTPUT_DIR" --video elementary "SOURCE"`,
	}
}

func validateBlurayVideoMode(mode string) error {
	switch mode {
	case blurayVideoElementary, blurayVideoM2TS, blurayVideoBoth:
		return nil
	default:
		return fmt.Errorf("--video must be elementary, m2ts or both (got %q)", mode)
	}
}

func cliSelectedBlurayPlaylist(ctx context.Context, raw, requested string) (bluraySource, blurayPlaylist, error) {
	source, err := resolveBluraySource(raw)
	if err != nil {
		return bluraySource{}, blurayPlaylist{}, err
	}
	playlists, err := discoverBlurayForSource(ctx, source)
	if err != nil {
		return bluraySource{}, blurayPlaylist{}, err
	}
	playlist, err := selectBlurayPlaylist(playlists, requested)
	return source, playlist, err
}

func blurayCLIInventory(playlist blurayPlaylist, streams []blurayStream) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Blu-ray playlist %05d — %s — %d chapters\n", playlist.Number, playlist.Duration(), len(playlist.Chapters))
	fmt.Fprintf(&b, "%-6s %-9s %-12s %-8s %-23s %s\n", "INDEX", "TYPE", "PID", "LANG", "CODEC", "TITLE")
	for _, s := range streams {
		pid := "-"
		if s.PID > 0 { pid = fmt.Sprintf("0x%04x", s.PID) }
		language := s.Language
		if language == "" { language = "und" }
		fmt.Fprintf(&b, "%-6d %-9s %-12s %-8s %-23s %s\n", s.Index, s.Kind, pid, language, s.Codec, bluRayStreamDisplayTitle(s))
	}
	return b.String()
}

// Reuse the verified ADVANCED MUX engine for a single M2TS with all compatible
// selected streams. No MKV intermediary is created. Chapter rows are omitted.
func cliBlurayMuxM2TS(ctx context.Context, tools toolPaths, source bluraySource, playlist blurayPlaylist, output string, indexes []int) error {
	available, err := probeBlurayPlaylist(ctx, tools, source, playlist)
	if err != nil { return err }
	selected, err := selectedBlurayStreams(available, indexes)
	if err != nil { return err }
	rows := make([]mergerStream, 0, len(selected))
	for _, s := range selected {
		rows = append(rows, mergerStream{
			Path: source.Input, BluRay: true, BluRayPlaylist: playlist.Number,
			Track: trackOption{Index:s.Index, Kind:s.Kind, Codec:s.Codec, Language:s.Language, Title:bluRayStreamDisplayTitle(s), Channels:s.Channels},
		})
	}
	if err := validateAdvancedM2TSStreams(rows); err != nil { return err }
	abs, err := filepath.Abs(output)
	if err != nil { return err }
	if err := os.MkdirAll(filepath.Dir(abs), 0755); err != nil { return err }
	return muxMerger(ctx, tools, rows, "", abs)
}
