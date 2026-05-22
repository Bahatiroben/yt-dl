package youtube

import (
    "errors"
    "regexp"
    "strings"
)

// ExtractVideoID extracts the video ID from various YouTube URL formats
func ExtractVideoID(url string) (string, error) {
    url = strings.TrimSpace(url)

    // Common patterns mostly found in YouTube URLs
    patterns := []*regexp.Regexp{
        // youtu.be/VIDEO_ID
        regexp.MustCompile(`youtu\.be/([a-zA-Z0-9_-]+)`),
        // youtube.com/watch?v=VIDEO_ID
        regexp.MustCompile(`[?&]v=([a-zA-Z0-9_-]+)`),
        // youtube.com/embed/VIDEO_ID
        regexp.MustCompile(`/embed/([a-zA-Z0-9_-]+)`),
        // youtube.com/shorts/VIDEO_ID
        regexp.MustCompile(`/shorts/([a-zA-Z0-9_-]+)`),
    }

    for _, pattern := range patterns {
        matches := pattern.FindStringSubmatch(url)
        if len(matches) > 1 {
            id := matches[1]
            if len(id) >= 11 {
				// TODO: Check frequently if YouTube changes their ID format in the future
                return id[:11], nil // YouTube IDs are 11 characters (atleast at the moment)
            }
        }
    }

    return "", errors.New("could not extract video ID from URL")
}