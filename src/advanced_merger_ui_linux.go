//go:build linux && !cli

package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func (g *linuxGUI) buildAdvancedMerger() fyne.CanvasObject {
	var streams []mergerStream
	var checks []*widget.Check
	list := container.NewVBox()
	chapter := widget.NewEntry()
	chapter.SetPlaceHolder("Optional chapter override: MKV or FFMETADATA1")
	output := widget.NewEntry()
	output.SetText(g.outputEntry.Text)
	name := widget.NewEntry()
	name.SetText("merged.mkv")
	status := widget.NewLabel("Choose files, then select the streams to include.")
	status.Wrapping = fyne.TextWrapWord
	activity := widget.NewProgressBarInfinite()
	activity.Stop()
	activity.Hide()
	var controls []fyne.Disableable
	var baseControlCount int
	var cancel context.CancelFunc
	cancelBtn := widget.NewButton("Cancel", func() {
		if cancel != nil {
			cancel()
		}
	})
	cancelBtn.Disable()
	busy := false
	run := func(label string, work func(context.Context) (func(), error)) {
		if busy {
			return
		}
		busy = true
		for _, c := range controls {
			c.Disable()
		}
		cancelBtn.Enable()
		status.SetText(label)
		activity.Show()
		activity.Start()
		ctx, c := context.WithCancel(context.Background())
		cancel = c
		go func() {
			done, err := work(ctx)
			c()
			fyne.Do(func() {
				busy = false
				cancel = nil
				for _, control := range controls {
					control.Enable()
				}
				cancelBtn.Disable()
				activity.Stop()
				activity.Hide()
				if err != nil {
					status.SetText(err.Error())
					dialog.ShowError(err, g.window)
				} else if done != nil {
					done()
				}
			})
		}()
	}
	addPaths := func(kind string, paths []string, playlistNumber int) {
		if len(paths) == 0 {
			return
		}
		run("Reading streams…", func(ctx context.Context) (func(), error) {
			tools, err := mergerTools(ctx)
			if err != nil {
				return nil, err
			}
			var added []mergerStream
			for _, p := range paths {
				var ss []mergerStream
				if kind == "bluray" {
					ss, err = probeMergerBluRay(ctx, p, "all", playlistNumber)
				} else {
					ss, err = probeMergerFile(ctx, tools.ffprobe, p, kind)
				}
				if err != nil {
					return nil, err
				}
				added = append(added, ss...)
			}
			return func() {
				for _, s := range added {
					duplicate := false
					for _, existing := range streams {
						if existing.Path == s.Path && existing.Track.Index == s.Track.Index && existing.BluRay == s.BluRay && existing.BluRayPlaylist == s.BluRayPlaylist {
							duplicate = true
						}
					}
					if duplicate {
						continue
					}
					streams = append(streams, s)
					check := widget.NewCheck(dvdSourceDisplayName(s.Path)+" — "+s.Track.Label(), nil)
					check.SetChecked(true)
					checks = append(checks, check)
					controls = append(controls, check)
					list.Add(check)
				}
				status.SetText(fmt.Sprintf("%d streams available", len(streams)))
			}, nil
		})
	}
	// Discover playlists before reading streams so looping fake-long titles are
	// never silently selected by ADVANCED. Require a user choice for ISO too.
	chooseBluRayPlaylist := func(path string) {
		run("Reading Blu-ray playlists…", func(ctx context.Context) (func(), error) {
			source, err := resolveBluraySource(path)
			if err != nil {
				return nil, err
			}
			playlists, err := discoverBlurayForSource(ctx, source)
			if err != nil {
				return nil, err
			}
			if len(playlists) == 0 {
				return nil, fmt.Errorf("Blu-ray contains no readable playlists")
			}
			return func() {
				labels := make([]string, 0, len(playlists))
				byLabel := make(map[string]int, len(playlists))
				for _, p := range playlists {
					label := fmt.Sprintf("Playlist %05d — %s — %d chapters", p.Number, p.Duration(), len(p.Chapters))
					labels = append(labels, label)
					byLabel[label] = p.Number
				}
				picker := widget.NewSelect(labels, nil)
				picker.PlaceHolder = "Choose a playlist (do not rely on longest)"
				info := widget.NewLabel("Choose the intended Blu-ray playlist. Some discs contain fake-long looping playlists.")
				info.Wrapping = fyne.TextWrapWord
				dialog.NewCustomConfirm("Choose ADVANCED Blu-ray playlist", "Use playlist", "Cancel",
					container.NewVBox(info, picker), func(ok bool) {
						if !ok {
							return
						}
						number, selected := byLabel[picker.Selected]
						if !selected {
							dialog.ShowInformation("Playlist required", "Select a playlist before adding Blu-ray streams.", g.window)
							return
						}
						addPaths("bluray", []string{path}, number)
					}, g.window).Show()
			}, nil
		})
	}
	add := func(kind string) {
		// A folder browser with checkboxes supports multiple files without external dialogs.
		// MEDIA can also add the currently browsed DVD/VIDEO_TS directory directly.
		dir := widget.NewEntry()
		home, _ := os.UserHomeDir()
		dir.SetText(home)
		files := container.NewVBox()
		selected := map[string]bool{}
		var refresh func()
		refresh = func() {
			entries, err := os.ReadDir(dir.Text)
			if err != nil {
				dialog.ShowError(err, g.window)
				return
			}
			files.Objects = nil
			files.Add(widget.NewButton("Parent folder", func() { dir.SetText(filepath.Dir(dir.Text)); refresh() }))
			for _, entry := range entries {
				entry := entry
				p := filepath.Join(dir.Text, entry.Name())
				if entry.IsDir() {
					files.Add(widget.NewButton(entry.Name()+"/", func() { dir.SetText(p); refresh() }))
				} else {
					check := widget.NewCheck(entry.Name(), func(on bool) { selected[p] = on })
					check.SetChecked(selected[p])
					files.Add(check)
				}
			}
			files.Refresh()
		}
		refresh()
		scroll := container.NewVScroll(files)
		scroll.SetMinSize(fyne.NewSize(600, 320))
		var chooser *dialog.ConfirmDialog
		actions := container.NewHBox(widget.NewButton("Open folder", refresh))
		if kind == "all" || kind == "bluray" {
			folderLabel := "Add current folder as DVD / VIDEO_TS"
			if kind == "bluray" {
				folderLabel = "Add current folder as Blu-ray / BDMV"
			}
			actions.Add(widget.NewButton(folderLabel, func() {
				if kind == "bluray" {
					if _, err := resolveBluraySource(dir.Text); err != nil {
						dialog.ShowError(err, g.window)
						return
					}
				} else if _, err := resolveDVDSource(dir.Text); err != nil {
					dialog.ShowError(err, g.window)
					return
				}
				if chooser != nil {
					chooser.Hide()
				}
				if kind == "bluray" {
					chooseBluRayPlaylist(dir.Text)
				} else {
					addPaths(kind, []string{dir.Text}, -1)
				}
			}))
		}
		chooser = dialog.NewCustomConfirm("Choose "+kind+" files", "Add files", "Cancel", container.NewBorder(container.NewBorder(nil, nil, nil, actions, dir), nil, nil, nil, scroll), func(ok bool) {
			if !ok {
				return
			}
			var paths []string
			for p, on := range selected {
				if on {
					paths = append(paths, p)
				}
			}
			sort.Strings(paths)
			if kind == "bluray" {
				if len(paths) != 1 {
					dialog.ShowInformation("Choose one Blu-ray", "Add one Blu-ray source at a time so its playlist can be selected.", g.window)
					return
				}
				chooseBluRayPlaylist(paths[0])
				return
			}
			addPaths(kind, paths, -1)
		}, g.window)
		chooser.Resize(fyne.NewSize(680, 460))
		chooser.Show()
	}
	movies := widget.NewButton("MEDIA(All streams included)", func() { add("all") })
	bluRay := widget.NewButton("BLU-RAY BDMV / ISO", func() { add("bluray") })
	drive := widget.NewButton("DVD DRIVE", func() {
		drives, err := listPhysicalDVDDrives()
		if err != nil {
			dialog.ShowError(err, g.window)
			return
		}
		if len(drives) == 0 {
			dialog.ShowInformation("DVD Drive", "No physical optical DVD/CD-ROM drive was detected.", g.window)
			return
		}
		options := make([]string, 0, len(drives))
		byLabel := make(map[string]string, len(drives))
		for _, dvd := range drives {
			label := dvd.Label + " — " + dvd.Input
			options = append(options, label)
			byLabel[label] = dvd.Input
		}
		selected := options[0]
		picker := widget.NewSelect(options, func(value string) { selected = value })
		picker.SetSelected(selected)
		dialog.NewCustomConfirm("Choose DVD Drive", "Add DVD", "Cancel", container.NewVBox(
			widget.NewLabel("The longest readable DVD title will be added as selectable streams."),
			picker,
		), func(ok bool) {
			if !ok {
				return
			}
			path := byLabel[selected]
			run("Reading DVD streams…", func(ctx context.Context) (func(), error) {
				tools, err := mergerTools(ctx)
				if err != nil {
					return nil, err
				}
				added, err := probeMergerFile(ctx, tools.ffprobe, path, "all")
				if err != nil {
					return nil, err
				}
				return func() {
					for _, s := range added {
						duplicate := false
						for _, existing := range streams {
							if existing.Path == s.Path && existing.Track.Index == s.Track.Index {
								duplicate = true
							}
						}
						if duplicate {
							continue
						}
						streams = append(streams, s)
						check := widget.NewCheck(dvdSourceDisplayName(s.Path)+" — "+s.Track.Label(), nil)
						check.SetChecked(true)
						checks = append(checks, check)
						controls = append(controls, check)
						list.Add(check)
					}
					status.SetText(fmt.Sprintf("%d streams available", len(streams)))
				}, nil
			})
		}, g.window).Show()
	})
	audio := widget.NewButton("ADD AUDIO(Only audio streams will be included)", func() { add("audio") })
	subs := widget.NewButton("ADD SUBTITLE(Only subtitle streams will be Included)", func() { add("subtitle") })
	chapters := widget.NewButton("ADD CHAPTER .txt FILE(FFMETADATA1 Format)", func() {
		dialog.ShowFileOpen(func(r fyne.URIReadCloser, err error) {
			if err != nil {
				dialog.ShowError(err, g.window)
				return
			}
			if r == nil {
				return
			}
			path := r.URI().Path()
			r.Close()
			run("Checking chapters…", func(ctx context.Context) (func(), error) {
				tools, err := mergerTools(ctx)
				if err != nil {
					return nil, err
				}
				if err = validateMergerChapters(ctx, tools.ffprobe, path); err != nil {
					return nil, err
				}
				return func() { chapter.SetText(path); status.SetText("Chapter file ready") }, nil
			})
		}, g.window)
	})
	folder := widget.NewButton("CHOOSE OUTPUT FOLDER", func() {
		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil {
				dialog.ShowError(err, g.window)
			} else if uri != nil {
				output.SetText(uri.Path())
			}
		}, g.window)
	})
	demux := widget.NewButton("DEMUX", func() {
		var chosen []mergerStream
		hasVideo := false
		for i, stream := range streams {
			if !checks[i].Checked {
				continue
			}
			chosen = append(chosen, stream)
			if stream.Track.Kind == "video" {
				hasVideo = true
			}
		}
		bluRay := selectedBluRayMergerStreams(chosen)
		videoMode := "mpeg2"
		options := container.NewVBox()
		var videoFormat *widget.Select
		if bluRay {
			videoMode = blurayVideoElementary
			if hasVideo {
				videoFormat = widget.NewSelect([]string{
					"Elementary video (codec-native .h264 / .hevc / .mpeg2)",
					"Video in M2TS transport stream (.m2ts)",
					"Both elementary video and M2TS",
				}, nil)
				videoFormat.SetSelected("Elementary video (codec-native .h264 / .hevc / .mpeg2)")
				options.Add(widget.NewLabel("Blu-ray video export format"))
				options.Add(videoFormat)
			} else {
				options.Add(widget.NewLabel("No Blu-ray video selected — exporting selected audio, subtitles and chapters only."))
			}
		} else if hasVideo {
			videoFormat = widget.NewSelect([]string{"MPEG2 elementary video (.mpeg2)", "VOB video (.VOB)"}, nil)
			videoFormat.SetSelected("MPEG2 elementary video (.mpeg2)")
			options.Add(widget.NewLabel("MPEG-2 video export format (DVD sources)"))
			options.Add(videoFormat)
		} else {
			options.Add(widget.NewLabel("No video stream selected — exporting selected audio, subtitles and chapters only."))
		}
		dialog.NewCustomConfirm("Demux selected streams", "Demux", "Cancel", options, func(ok bool) {
			if !ok {
				return
			}
			if bluRay && hasVideo {
				switch videoFormat.Selected {
				case "Video in M2TS transport stream (.m2ts)":
					videoMode = blurayVideoM2TS
				case "Both elementary video and M2TS":
					videoMode = blurayVideoBoth
				}
			} else if !bluRay && hasVideo && videoFormat.Selected == "VOB video (.VOB)" {
				videoMode = "vob"
			}
			dir := output.Text
			run("Demuxing selected streams…", func(ctx context.Context) (func(), error) {
				if err := validateOutputDir(dir); err != nil {
					return nil, err
				}
				tools, err := mergerTools(ctx)
				if err != nil {
					return nil, err
				}
				final, err := demuxMerger(ctx, tools, chosen, dir, videoMode)
				if err != nil {
					return nil, err
				}
				return func() { status.SetText("Demux complete: " + final) }, nil
			})
		}, g.window).Show()
	})
	mux := widget.NewButton("MUX", func() {
		var chosen []mergerStream
		for i, s := range streams {
			if checks[i].Checked {
				chosen = append(chosen, s)
			}
		}
		filename := strings.TrimSpace(name.Text)
		if filepath.Base(filename) != filename || !strings.HasSuffix(strings.ToLower(filename), ".mkv") {
			dialog.ShowInformation("Output filename", "Enter a filename ending in .mkv, without folders.", g.window)
			return
		}
		dir, chap := output.Text, chapter.Text
		run("Muxing selected streams…", func(ctx context.Context) (func(), error) {
			if err := validateOutputDir(dir); err != nil {
				return nil, err
			}
			tools, err := mergerTools(ctx)
			if err != nil {
				return nil, err
			}
			path := filepath.Join(dir, filename)
			if err = muxMerger(ctx, tools, chosen, chap, path); err != nil {
				return nil, err
			}
			return func() { status.SetText("Completed: " + path) }, nil
		})
	})
	mux.Importance = widget.HighImportance
	clear := widget.NewButton("Clear streams", func() {
		streams = nil
		checks = nil
		controls = controls[:baseControlCount]
		list.Objects = nil
		list.Refresh()
		status.SetText("Choose files to add streams.")
	})
	controls = []fyne.Disableable{movies, bluRay, drive, audio, subs, chapters, folder, demux, mux, clear, chapter, output, name}
	baseControlCount = len(controls)
	scroll := container.NewVScroll(list)
	return container.NewBorder(container.NewVBox(container.NewGridWithColumns(2, movies, drive, audio, subs, bluRay, widget.NewLabel("Blu-ray: choose playlist / lossless LPCM→FLAC")), widget.NewLabel("Select Streams — choose one chapter set, or use the chapter override below")), container.NewVBox(clear, container.NewBorder(nil, nil, nil, chapters, chapter), container.NewBorder(nil, nil, nil, folder, output), container.NewBorder(nil, nil, widget.NewLabel("Output filename"), nil, name), status, activity, container.NewHBox(demux, mux, cancelBtn)), nil, nil, scroll)
}
