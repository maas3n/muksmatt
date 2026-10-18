//go:build windows

package main

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"unsafe"
)

const wmShowRemuxBluRayPlaylist = WM_APP + 46

// The disk scan runs on a worker, but Win32 dialog controls must be created
// on the UI thread. Publish this request only after startAsync has cleared the
// busy state, so the normal UI reset cannot re-enable a modal dialog's owner.
var remuxBluRayPlaylistRequest struct {
	sync.Mutex
	source    string
	playlists []blurayPlaylist
}

func queueRemuxBluRayPlaylistChooser(source string, playlists []blurayPlaylist) {
	remuxBluRayPlaylistRequest.Lock()
	remuxBluRayPlaylistRequest.source = source
	remuxBluRayPlaylistRequest.playlists = append([]blurayPlaylist(nil), playlists...)
	remuxBluRayPlaylistRequest.Unlock()
}

func postRemuxBluRayPlaylistChooserIfPending() {
	remuxBluRayPlaylistRequest.Lock()
	pending := len(remuxBluRayPlaylistRequest.playlists) != 0
	remuxBluRayPlaylistRequest.Unlock()
	if pending {
		procPostMessageW.Call(app.hwnd, wmShowRemuxBluRayPlaylist, 0, 0)
	}
}

// Called by the main window procedure, never from the scan goroutine.
func showPendingRemuxBluRayPlaylistChooser() {
	remuxBluRayPlaylistRequest.Lock()
	source := remuxBluRayPlaylistRequest.source
	playlists := remuxBluRayPlaylistRequest.playlists
	remuxBluRayPlaylistRequest.source = ""
	remuxBluRayPlaylistRequest.playlists = nil
	remuxBluRayPlaylistRequest.Unlock()
	if len(playlists) == 0 { return }
	showWindowsBluRayPlaylistChooser("Choose REMUX/DEMUX Blu-ray playlist", playlists, func(number int) {
		current, err := currentSource()
		if err != nil || !windowsBlurayMode() || !strings.EqualFold(filepath.Clean(current), filepath.Clean(source)) {
			messageBox(app.hwnd, "Blu-ray source changed", "The source changed. Scan again before selecting tracks.", MB_OK|MB_ICONWARNING)
			return
		}
		app.titlesMu.Lock()
		valid := strings.EqualFold(filepath.Clean(app.titlesSource), filepath.Clean(source))
		_, selectedErr := selectedRemuxBluRayPlaylist(app.titles, number)
		index := -1
		if valid && selectedErr == nil {
			for i, title := range app.titles {
				if title.Number == number { index = i; break }
			}
		}
		app.titlesMu.Unlock()
		if index < 0 {
			messageBox(app.hwnd, "Playlist unavailable", "The selected playlist is no longer in the scan. Scan again.", MB_OK|MB_ICONWARNING)
			return
		}
		procSendMessageW.Call(app.titleCombo, CB_SETCURSEL, uintptr(index), 0)
		clearWindowsTrackSelection()
		startAsync(fmt.Sprintf("Reading Blu-ray playlist %05d tracks…", number), showMetadata)
	})
}

// Unlike the DVD scan, Blu-ray scanning never selects the longest playlist.
// The only path into the stream chooser is an explicit human selection.
func scanWindowsRemuxBluRay(ctx context.Context) error {
	src, err := currentSource()
	if err != nil { return err }
	source, err := resolveBluraySource(src)
	if err != nil { return err }
	playlists, err := discoverBlurayForSource(ctx, source)
	if err != nil { return err }
	if len(playlists) == 0 {
		return errors.New("Blu-ray contains no readable playlists")
	}
	if err := ctx.Err(); err != nil { return err }
	titles := make([]titleInfo, 0, len(playlists))
	for _, p := range playlists {
		titles = append(titles, titleInfo{Number: p.Number, Duration: p.Duration()})
	}
	app.titlesMu.Lock()
	app.titles = titles
	app.titlesSource = source.Input
	app.titlesMu.Unlock()
	clearWindowsTrackSelection()
	procSendMessageW.Call(app.titleCombo, CB_RESETCONTENT, 0, 0)
	for _, title := range titles {
		label := remuxBluRayPlaylistLabel(title.Number, title.Duration)
		procSendMessageW.Call(app.titleCombo, CB_ADDSTRING, 0, uintptr(unsafe.Pointer(utf16Ptr(label))))
	}
	// -1 means no selection; never preselect the fake-long '00152' playlist.
	procSendMessageW.Call(app.titleCombo, CB_SETCURSEL, ^uintptr(0), 0)
	setProgress(1)
	setStatus(fmt.Sprintf("Found %d Blu-ray playlists. Choose one to inspect its streams.", len(playlists)))
	queueRemuxBluRayPlaylistChooser(source.Input, playlists)
	return nil
}
