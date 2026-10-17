//go:build windows

package main

import (
	"syscall"
	"unsafe"
)

const (
	idAdvancedMetaLang = 7391
	idAdvancedMetaTitle = 7392
	idAdvancedMetaApply = 7393
	idAdvancedMetaCancel = 7394
)

type advancedMetadataDialogState struct {
	hwnd, language, title uintptr
	onApply func(string, string) error
}

var advancedMetadataDialog *advancedMetadataDialogState

func showWindowsAdvancedMetadataEditor(language, title string, onApply func(string, string) error) {
	if advancedMetadataDialog != nil { return }
	hInstance, _, _ := procGetModuleHandleW.Call(0)
	className := utf16Ptr("muKsMaTTAdvancedMetadataEditor")
	cursor, _, _ := procLoadCursorW.Call(0, 32512)
	wc := WNDCLASSEX{
		CbSize: uint32(unsafe.Sizeof(WNDCLASSEX{})),
		LpfnWndProc: syscall.NewCallback(advancedMetadataWindowProc),
		HInstance: hInstance,
		HCursor: cursor,
		HbrBackground: 6,
		LpszClassName: className,
	}
	procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))
	dpi := windowDPI(app.hwnd)
	sc := func(v int32) uintptr { return uintptr(scale96(v, dpi)) }
	hwnd, _, _ := procCreateWindowExW.Call(WS_EX_CONTROLPARENT,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(utf16Ptr("ADVANCED — Edit track metadata"))),
		uintptr(WS_OVERLAPPED|WS_CAPTION|WS_SYSMENU),
		CW_USEDEFAULT, CW_USEDEFAULT, sc(615), sc(270),
		app.hwnd, 0, hInstance, 0)
	if hwnd == 0 {
		messageBox(app.hwnd, "Track metadata", "Cannot open the metadata editor.", MB_OK|MB_ICONERROR)
		return
	}
	add := func(class, text string, style uint32, x,y,w,h int32, id int) uintptr {
		control, _, _ := procCreateWindowExW.Call(0, uintptr(unsafe.Pointer(utf16Ptr(class))),
			uintptr(unsafe.Pointer(utf16Ptr(text))), uintptr(WS_CHILD|WS_VISIBLE|style),
			sc(x),sc(y),sc(w),sc(h), hwnd, uintptr(id), hInstance, 0)
		if control != 0 { procSendMessageW.Call(control, WM_SETFONT, app.bodyFont, 1) }
		return control
	}
	add("STATIC", "Language: ISO 639-2, e.g. eng / nor / spa (blank clears the tag)",
		0, 18, 14, 570, 25, 0)
	langEdit := add("EDIT", language, WS_BORDER|ES_AUTOHSCROLL|WS_TABSTOP, 18, 43, 565, 29, idAdvancedMetaLang)
	add("STATIC", "Track title (blank clears the tag)", 0, 18, 83, 565, 25, 0)
	titleEdit := add("EDIT", title, WS_BORDER|ES_AUTOHSCROLL|WS_TABSTOP, 18, 110, 565, 29, idAdvancedMetaTitle)
	add("BUTTON", "Apply", BS_DEFPUSHBUTTON|WS_TABSTOP, 350, 161, 112, 32, idAdvancedMetaApply)
	add("BUTTON", "Cancel", BS_PUSHBUTTON|WS_TABSTOP, 471, 161, 112, 32, idAdvancedMetaCancel)
	advancedMetadataDialog = &advancedMetadataDialogState{hwnd: hwnd, language: langEdit, title: titleEdit, onApply: onApply}
	procEnableWindow.Call(app.hwnd, 0)
	procShowWindow.Call(hwnd, SW_SHOW)
	procUpdateWindow.Call(hwnd)
}

func advancedMetadataWindowProc(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case WM_COMMAND:
		d := advancedMetadataDialog
		if d == nil || d.hwnd != hwnd { break }
		switch int(wParam & 0xffff) {
		case idAdvancedMetaApply:
			if d.onApply != nil {
				if err := d.onApply(getText(d.language), getText(d.title)); err != nil {
					messageBox(hwnd, "Invalid track metadata", err.Error(), MB_OK|MB_ICONWARNING)
					return 0
				}
			}
			procDestroyWindow.Call(hwnd)
			return 0
		case idAdvancedMetaCancel:
			procDestroyWindow.Call(hwnd)
			return 0
		}
	case WM_CLOSE:
		procDestroyWindow.Call(hwnd)
		return 0
	case WM_DESTROY:
		if advancedMetadataDialog != nil && advancedMetadataDialog.hwnd == hwnd {
			advancedMetadataDialog = nil
			procEnableWindow.Call(app.hwnd, 1)
		}
		return 0
	}
	result, _, _ := procDefWindowProcW.Call(hwnd, uintptr(msg), wParam, lParam)
	return result
}
