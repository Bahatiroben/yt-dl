package youtube

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
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
func ExtractPlayerResponse(html string) (map[string]interface{}, error) {
    // More robust patterns used by many scrapers in 2026
    patterns := []*regexp.Regexp{
        regexp.MustCompile(`var ytInitialPlayerResponse\s*=\s*(\{.+?\});\s*var`),
        regexp.MustCompile(`var ytInitialPlayerResponse\s*=\s*(\{[\s\S]*?\});\s*var`),
        regexp.MustCompile(`(?s)ytInitialPlayerResponse\s*=\s*(\{.+?\})\s*;\s*var`),
        regexp.MustCompile(`"playerResponse"\s*:\s*(\{.+?\})\s*,"`),
    }

    for i, re := range patterns {
        matches := re.FindStringSubmatch(html)
        if len(matches) > 1 {
            jsonStr := matches[1]
            fmt.Printf("✅ Pattern %d matched, attempting to parse JSON...\n", i+1)

            // Try to parse
            var playerResponse map[string]interface{}
            err := json.Unmarshal([]byte(jsonStr), &playerResponse)
            if err == nil {
                fmt.Println("✅ Successfully parsed playerResponse JSON!")
                return playerResponse, nil
            }
            fmt.Printf("   JSON parsing failed: %v\n", err)
        }
    }

    // === DEBUG INFO ===
    fmt.Println("\n🔍 Advanced Debug:")
    if idx := strings.Index(html, "ytInitialPlayerResponse"); idx != -1 {
        fmt.Println("✅ 'ytInitialPlayerResponse' FOUND in HTML")
        start := max(0, idx-120)
        end := min(len(html), idx+600)
        fmt.Println("Context snippet:")
        fmt.Println(html[start:end])
    } else {
        fmt.Println("❌ 'ytInitialPlayerResponse' NOT found in the entire page")
    }

    // Save debug file
    os.WriteFile("debug_page.html", []byte(html), 0644)
    fmt.Println("💾 Full page saved to debug_page.html")

    return nil, errors.New("could not extract ytInitialPlayerResponse")
}

func max(a, b int) int { if a > b { return a }; return b }
func min(a, b int) int { if a < b { return a }; return b }

// ExtractVideoMetadata - improved for current Innertube API
func ExtractVideoMetadata(playerResponse map[string]interface{}) (Video, error) {
    video := Video{}

    fmt.Println("🔍 Extracting metadata...")

    // Primary location: videoDetails
    if videoDetails, ok := playerResponse["videoDetails"].(map[string]interface{}); ok {
        fmt.Println("✅ Found videoDetails")

        if title, ok := videoDetails["title"].(string); ok && title != "" {
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

    // Fallback: check root level
    if video.Title == "" {
        if title, ok := playerResponse["title"].(string); ok {
            video.Title = title
        }
    }

    if video.Title == "" {
        // Print available keys for debugging
        fmt.Println("🔍 Available keys in playerResponse:")
        for k := range playerResponse {
            fmt.Printf("   • %s\n", k)
        }
        return video, errors.New("could not extract video title")
    }

    fmt.Printf("✅ Extracted Title: %s\n", video.Title)
    return video, nil
}

// ExtractFormats - improved with better debugging
func ExtractFormats(playerResponse map[string]interface{}) ([]Format, error) {
    var formats []Format

    // Check playability status first
    if playability, ok := playerResponse["playabilityStatus"].(map[string]interface{}); ok {
        if status, ok := playability["status"].(string); ok {
            fmt.Printf("🎮 Playability Status: %s\n", status)
            if status != "OK" {
                if reason, ok := playability["reason"].(string); ok {
                    fmt.Printf("❌ Reason: %s\n", reason)
                }
                return nil, fmt.Errorf("video not playable: %s", status)
            }
        }
    }

    // Try to find streamingData
    streamingData, ok := playerResponse["streamingData"].(map[string]interface{})
    if !ok {
        fmt.Println("🔍 streamingData not found at root. Available keys:")
        for k := range playerResponse {
            fmt.Printf("   • %s\n", k)
        }
        return nil, errors.New("streamingData not found")
    }

    fmt.Println("✅ Found streamingData")

    // Process formats
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

                    if mimeType, ok := formatMap["mimeType"].(string); ok {
                        if strings.Contains(mimeType, "audio/") {
                            format.HasVideo = false
                        } else if strings.Contains(mimeType, "video/") && !strings.Contains(mimeType, "audio") {
                            format.HasAudio = false
                        }
                    }

                    formats = append(formats, format)
                }
            }
        }
    }

    if regular, exists := streamingData["formats"]; exists {
        processFormats(regular)
    }
    if adaptive, exists := streamingData["adaptiveFormats"]; exists {
        processFormats(adaptive)
    }

    if len(formats) == 0 {
        return nil, errors.New("no formats found in streamingData")
    }

    return formats, nil
}

// SelectBestFormat selects the best format preferring direct downloadable URLs
func SelectBestFormat(formats []Format) (*Format, error) {
	var best *Format

	for i := range formats {
		f := &formats[i]

		// Skip formats without URL for now (they need deciphering)
		if f.URL == "" {
			continue
		}

		// Priority 1: MP4 with both video + audio
		if strings.Contains(f.MimeType, "mp4") && f.HasVideo && f.HasAudio {
			if best == nil || f.Bitrate > best.Bitrate {
				best = f
			}
		}
	}

	// Priority 2: Any MP4 with direct URL
	if best == nil {
		for i := range formats {
			f := &formats[i]
			if f.URL == "" {
				continue
			}
			if strings.Contains(f.MimeType, "mp4") {
				if best == nil || f.Bitrate > best.Bitrate {
					best = f
				}
			}
		}
	}

	// Priority 3: Any format with direct URL (last resort)
	if best == nil {
		for i := range formats {
			f := &formats[i]
			if f.URL != "" {
				if best == nil || f.Bitrate > best.Bitrate {
					best = f
				}
			}
		}
	}

	if best == nil {
		return nil, errors.New("no downloadable format found (all require deciphering)")
	}

	return best, nil
}
