//go:build linux

package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Linux optical drives must be real CD/DVD/Blu-ray class-5 block devices.
// Do not accept arbitrary block devices (for example /dev/sda).
func resolvePhysicalBlurayDrive(raw string) (bluraySource, bool, error) {
	path := strings.TrimSpace(strings.Trim(raw, "\""))
	if !strings.HasPrefix(path, "/dev/") {
		return bluraySource{}, false, nil
	}
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		resolved = path
	}
	if !linuxOpticalDevice(resolved) {
		return bluraySource{}, false, fmt.Errorf("not an accessible optical Blu-ray drive: %s", path)
	}
	return bluraySource{
		Kind: bluraySourcePhysicalDrive,
		Input: resolved,
		Label: "Blu-ray drive " + resolved,
		BaseName: "Blu-ray-" + filepath.Base(resolved),
	}, true, nil
}
