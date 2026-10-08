//go:build windows

package main

import (
	"fmt"
	"strings"
)

// Reuse the proven kernel32 drive type check from the DVD drive enumeration,
// but return a distinct Blu-ray source which never touches DVD processing.
func resolvePhysicalBlurayDrive(raw string) (bluraySource, bool, error) {
	input := strings.TrimSpace(strings.Trim(raw, "\""))
	letter, isDrive := windowsDriveLetter(input)
	if !isDrive {
		return bluraySource{}, false, nil
	}
	if windowsDriveType(letter) != driveCDROM {
		return bluraySource{}, false, fmt.Errorf("%c: is not an optical drive", letter)
	}
	return bluraySource{
		Kind: bluraySourcePhysicalDrive,
		Input: windowsDriveRoot(letter),
		Label: fmt.Sprintf("Blu-ray drive %c:", letter),
		BaseName: fmt.Sprintf("Blu-ray-%c", letter),
	}, true, nil
}
