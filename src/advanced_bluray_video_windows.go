//go:build windows

package main

import (
	"syscall"
	"unsafe"
)

const (
	idAdvancedBluRayVideoCombo = 7380
	idAdvancedBluRayVideoUse = 7381
	idAdvancedBluRayVideoCancel = 7382
)

type windowsAdvancedBluRayVideoDialog struct {
	hwnd uintptr
	combo uintptr
	onChoose func(string)
}

var advancedBluRayVideoDialog *windowsAdvancedBluRayVideoDialog

// ADVANCED Blu-ray video streams may be exported as their elementary codec
// format, transported in actual M2TS packets, or both. DVD MPEG-2/VOB
// choices must not be shown for a Blu-ray source.
func showWindowsAdvancedBluRayVideoChooser(onChoose func(string)) {
	if advancedBluRayVideoDialog != nil {
		return
	}
	hInstance, _, _ := procGetModuleHandleW.Call(0)
	className := utf16Ptr("muKsMaTTAdvancedBluRayVideoChooser")
	cursor, _, _ := procLoadCursorW.Call(0, 32512)
	wc := WNDCLASSEX{
		CbSize: uint32(unsafe.Sizeof(WNDCLASSEX{})),
		LpfnWndProc: syscall.NewCallback(windowsAdvancedBluRayVideoWindowProc),
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
		uintptr(unsafe.Pointer(utf16Ptr("Blu-ray video export — ADVANCED"))),
		uintptr(WS_OVERLAPPED|WS_CAPTION|WS_SYSMENU),
		CW_USEDEFAULT, CW_USEDEFAULT, sc(660), sc(225),
		app.hwnd, 0, hInstance, 0)
	if hwnd == 0 {
		messageBox(app.hwnd, "Blu-ray video", "Could not create the Blu-ray video format selector.", MB_OK|MB_ICONERROR)
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
	add("STATIC", "Select the Blu-ray video output. Audio, subtitles and chapters stay separate.",
		0, 18, 16, 615, 32, 0)
	combo := add("COMBOBOX", "", CBS_DROPDOWNLIST|WS_TABSTOP|WS_VSCROLL,
		18, 57, 615, 240, idAdvancedBluRayVideoCombo)
	for _, label := range []string{
		"Elementary video (.h264 / .hevc / codec-native)",
		"Blu-ray M2TS transport stream (.m2ts)",
		"Both elementary video and M2TS",
	} {
		procSendMessageW.Call(combo, CB_ADDSTRING, 0, uintptr(unsafe.Pointer(utf16Ptr(label))))
	}
	procSendMessageW.Call(combo, CB_SETCURSEL, 0, 0)
	add("BUTTON", "Demux", BS_DEFPUSHBUTTON|WS_TABSTOP,
		389, 112, 125, 32, idAdvancedBluRayVideoUse)
	add("BUTTON", "Cancel", BS_PUSHBUTTON|WS_TABSTOP,
		524, 112, 109, 32, idAdvancedBluRayVideoCancel)
	advancedBluRayVideoDialog = &windowsAdvancedBluRayVideoDialog{hwnd: hwnd, combo: combo, onChoose: onChoose}
	procEnableWindow.Call(app.hwnd, 0)
	procShowWindow.Call(hwnd, SW_SHOW)
	procUpdateWindow.Call(hwnd)
}

func windowsAdvancedBluRayVideoWindowProc(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case WM_COMMAND:
		dlg := advancedBluRayVideoDialog
		if dlg == nil || dlg.hwnd != hwnd {
			break
		}
		switch int(wParam & 0xffff) {
		case idAdvancedBluRayVideoUse:
			idx, _, _ := procSendMessageW.Call(dlg.combo, CB_GETCURSEL, 0, 0)
			modes := []string{blurayVideoElementary, blurayVideoM2TS, blurayVideoBoth}
			if idx >= uintptr(len(modes)) {
				messageBox(hwnd, "Blu-ray video", "Choose a Blu-ray video export format.", MB_OK|MB_ICONWARNING)
				return 0
			}
			callback, mode := dlg.onChoose, modes[idx]
			procDestroyWindow.Call(hwnd)
			if callback != nil {
				callback(mode)
			}
			return 0
		case idAdvancedBluRayVideoCancel:
			procDestroyWindow.Call(hwnd)
			return 0
		}
	case WM_CLOSE:
		procDestroyWindow.Call(hwnd)
		return 0
	case WM_DESTROY:
		if advancedBluRayVideoDialog != nil && advancedBluRayVideoDialog.hwnd == hwnd {
			advancedBluRayVideoDialog = nil
			procEnableWindow.Call(app.hwnd, 1)
		}
		return 0
	}
	r, _, _ := procDefWindowProcW.Call(hwnd, uintptr(msg), wParam, lParam)
	return r
}
