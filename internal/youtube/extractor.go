package youtube

import (
	"encoding/json"
	"errors"
	"fmt"
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

// ExtractPlayerResponse extracts the ytInitialPlayerResponse JSON from page HTML
func ExtractPlayerResponse(html string) (map[string]any, error) {
	// Regex to find ytInitialPlayerResponse
	// TODO: Keep an eye on YouTube's page structure changes, as this regex might need updates in the future
	re := regexp.MustCompile(`var ytInitialPlayerResponse\s*=\s*(\{.+?\});`)
	matches := re.FindStringSubmatch(html)

	if len(matches) < 2 {
		return nil, errors.New("could not find ytInitialPlayerResponse in page")
	}

	jsonStr := matches[1]

	// Parse JSON
	var playerResponse map[string]any
	err := json.Unmarshal([]byte(jsonStr), &playerResponse)
	if err != nil {
		return nil, errors.New("failed to parse ytInitialPlayerResponse JSON")
	}

	return playerResponse, nil
}

// ExtractVideoMetadata extracts title, author, length from player response
func ExtractVideoMetadata(playerResponse map[string]interface{}) (Video, error) {
	video := Video{}

	// Extract title
	if videoDetails, ok := playerResponse["videoDetails"].(map[string]interface{}); ok {
		if title, ok := videoDetails["title"].(string); ok {
			video.Title = title
		}
		if author, ok := videoDetails["author"].(string); ok {
			video.Author = author
		}
		if lengthStr, ok := videoDetails["lengthSeconds"].(string); ok {
			fmt.Sscanf(lengthStr, "%d", &video.Length)
		}
		if id, ok := videoDetails["videoId"].(string); ok {
			video.ID = id
		}
	}

	if video.Title == "" {
		return video, errors.New("could not extract video title")
	}

	return video, nil
}

// ExtractFormats extracts available formats from streamingData
func ExtractFormats(playerResponse map[string]interface{}) ([]Format, error) {
    var formats []Format

    streamingData, ok := playerResponse["streamingData"].(map[string]interface{})
    if !ok {
        return nil, errors.New("streamingData not found in player response")
    }

    // Helper function to process format array
    processFormats := func(formatList interface{}) {
        if formatsArray, ok := formatList.([]interface{}); ok {
            for _, f := range formatsArray {
                if formatMap, ok := f.(map[string]interface{}); ok {
                    format := Format{
                        HasVideo: true,
                        HasAudio: true,
                    }

                    if itag, ok := formatMap["itag"].(float64); ok {
                        format.Itag = int(itag)
                    }
                    if qualityLabel, ok := formatMap["qualityLabel"].(string); ok {
                        format.QualityLabel = qualityLabel
                    }
                    if mimeType, ok := formatMap["mimeType"].(string); ok {
                        format.MimeType = mimeType
                    }
                    if bitrate, ok := formatMap["bitrate"].(float64); ok {
                        format.Bitrate = int(bitrate)
                    }
                    if url, ok := formatMap["url"].(string); ok {
                        format.URL = url
                    }

                    // Check if it's audio only or video only
                    if mimeType, ok := formatMap["mimeType"].(string); ok {
                        if strings.Contains(mimeType, "audio/") {
                            format.HasVideo = false
                            format.HasAudio = true
                        } else if strings.Contains(mimeType, "video/") {
                            format.HasAudio = false // adaptive video usually has no audio
                        }
                    }

                    formats = append(formats, format)
                }
            }
        }
    }

    // Process both regular formats and adaptive formats
    if regularFormats, exists := streamingData["formats"]; exists {
        processFormats(regularFormats)
    }
    if adaptiveFormats, exists := streamingData["adaptiveFormats"]; exists {
        processFormats(adaptiveFormats)
    }

    if len(formats) == 0 {
        return nil, errors.New("no formats found")
    }

    return formats, nil
}
