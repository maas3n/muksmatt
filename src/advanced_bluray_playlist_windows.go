//go:build windows

package main

import (
	"fmt"
	"syscall"
	"unsafe"
)

const (
	idAdvancedBDPlaylistCombo = 7350
	idAdvancedBDPlaylistUse   = 7351
	idAdvancedBDPlaylistCancel = 7352
)

type windowsAdvancedBDPlaylistDialog struct {
	hwnd      uintptr
	combo     uintptr
	playlists []blurayPlaylist
	onChoose  func(int)
}

var advancedBDPlaylistDialog *windowsAdvancedBDPlaylistDialog

// The source has already been enumerated in a worker. Present every playlist
// (including possible decoys) so the user can choose by number, rather than
// silently running an arbitrary longest playlist through FFmpeg.
func showWindowsAdvancedBluRayPlaylistChooser(playlists []blurayPlaylist, onChoose func(int)) {
	showWindowsBluRayPlaylistChooser("Choose Blu-ray playlist — ADVANCED", playlists, onChoose)
}

// Reuse the same native selector in REMUX/DEMUX; neither workflow silently
// defaults to the longest (potentially looping) playlist.
func showWindowsBluRayPlaylistChooser(title string, playlists []blurayPlaylist, onChoose func(int)) {
	if advancedBDPlaylistDialog != nil || len(playlists) == 0 {
		return
	}
	hInstance, _, _ := procGetModuleHandleW.Call(0)
	className := utf16Ptr("muKsMaTTAdvancedBluRayPlaylistChooser")
	cursor, _, _ := procLoadCursorW.Call(0, 32512)
	wc := WNDCLASSEX{
		CbSize: uint32(unsafe.Sizeof(WNDCLASSEX{})),
		LpfnWndProc: syscall.NewCallback(windowsAdvancedBDPlaylistWindowProc),
		HInstance: hInstance,
		HCursor: cursor,
		HbrBackground: 6, // COLOR_WINDOW+1
		LpszClassName: className,
	}
	procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))
	dpi := windowDPI(app.hwnd)
	sc := func(n int32) uintptr { return uintptr(scale96(n, dpi)) }
	hwnd, _, _ := procCreateWindowExW.Call(WS_EX_CONTROLPARENT,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(utf16Ptr(title))),
		uintptr(WS_OVERLAPPED|WS_CAPTION|WS_SYSMENU),
		CW_USEDEFAULT, CW_USEDEFAULT, sc(650), sc(225),
		app.hwnd, 0, hInstance, 0)
	if hwnd == 0 {
		messageBox(app.hwnd, "Blu-ray playlist", "Could not create the playlist selector.", MB_OK|MB_ICONERROR)
		return
	}
	add := func(class, label string, style uint32, x, y, w, h int32, id int) uintptr {
		control, _, _ := procCreateWindowExW.Call(0,
			uintptr(unsafe.Pointer(utf16Ptr(class))),
			uintptr(unsafe.Pointer(utf16Ptr(label))),
			uintptr(WS_CHILD|WS_VISIBLE|style),
			sc(x), sc(y), sc(w), sc(h), hwnd, uintptr(id), hInstance, 0)
		if control != 0 {
			procSendMessageW.Call(control, WM_SETFONT, app.bodyFont, 1)
		}
		return control
	}
	add("STATIC", "Select the movie playlist explicitly; long playlists can be loops.",
		0, 18, 16, 605, 30, 0)
	combo := add("COMBOBOX", "", CBS_DROPDOWNLIST|WS_TABSTOP|WS_VSCROLL,
		18, 55, 605, 260, idAdvancedBDPlaylistCombo)
	for _, p := range playlists {
		label := fmt.Sprintf("Playlist %05d — %s — %d chapters", p.Number, p.Duration(), len(p.Chapters))
		procSendMessageW.Call(combo, CB_ADDSTRING, 0, uintptr(unsafe.Pointer(utf16Ptr(label))))
	}
	// No automatic CB_SETCURSEL: the user must choose a playlist.
	add("BUTTON", "Use playlist", BS_DEFPUSHBUTTON|WS_TABSTOP,
		375, 112, 135, 32, idAdvancedBDPlaylistUse)
	add("BUTTON", "Cancel", BS_PUSHBUTTON|WS_TABSTOP,
		520, 112, 103, 32, idAdvancedBDPlaylistCancel)
	advancedBDPlaylistDialog = &windowsAdvancedBDPlaylistDialog{
		hwnd: hwnd, combo: combo, playlists: append([]blurayPlaylist(nil), playlists...), onChoose: onChoose,
	}
	procEnableWindow.Call(app.hwnd, 0)
	procShowWindow.Call(hwnd, SW_SHOW)
	procUpdateWindow.Call(hwnd)
}

func windowsAdvancedBDPlaylistWindowProc(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case WM_COMMAND:
		dlg := advancedBDPlaylistDialog
		if dlg == nil || dlg.hwnd != hwnd {
			break
		}
		switch int(wParam & 0xffff) {
		case idAdvancedBDPlaylistUse:
			index, _, _ := procSendMessageW.Call(dlg.combo, CB_GETCURSEL, 0, 0)
			if index >= uintptr(len(dlg.playlists)) {
				messageBox(hwnd, "Playlist required", "Choose a Blu-ray playlist before adding streams.", MB_OK|MB_ICONWARNING)
				return 0
			}
			number, callback := dlg.playlists[index].Number, dlg.onChoose
			procDestroyWindow.Call(hwnd)
			if callback != nil {
				callback(number)
			}
			return 0
		case idAdvancedBDPlaylistCancel:
			procDestroyWindow.Call(hwnd)
			return 0
		}
	case WM_CLOSE:
		procDestroyWindow.Call(hwnd)
		return 0
	case WM_DESTROY:
		if advancedBDPlaylistDialog != nil && advancedBDPlaylistDialog.hwnd == hwnd {
			advancedBDPlaylistDialog = nil
			procEnableWindow.Call(app.hwnd, 1)
		}
		return 0
	}
	r, _, _ := procDefWindowProcW.Call(hwnd, uintptr(msg), wParam, lParam)
	return r
}
