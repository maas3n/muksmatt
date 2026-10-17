//go:build windows || linux

package main

import (
	"fmt"
	"strings"
)

func setAdvancedStreamMetadata(stream *mergerStream, language, title string) error {
	if stream == nil || (stream.Track.Kind != "audio" && stream.Track.Kind != "subtitle") {
		return fmt.Errorf("select one audio or subtitle stream to edit")
	}
	language = strings.ToLower(strings.TrimSpace(language))
	if language != "" {
		if len(language) != 3 {
			return fmt.Errorf("enter a three-letter ISO 639-2 language code (eng, nor, spa), or leave blank")
		}
		for _, r := range language {
			if r < 'a' || r > 'z' {
				return fmt.Errorf("language code must contain three letters")
			}
		}
	}
	title = strings.TrimSpace(title)
	if strings.ContainsAny(title, "\r\n") {
		return fmt.Errorf("track title must be a single line")
	}
	if len(title) > 256 {
		return fmt.Errorf("track title must be 256 characters or fewer")
	}
	stream.Track.Language = language
	stream.Track.Title = title
	stream.MetadataEdited = true
	return nil
}
