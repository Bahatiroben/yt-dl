package youtube

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type Client struct {
	httpClient *http.Client
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) GetVideo(url string) (*Video, error) {
	videoID, err := ExtractVideoID(url)
	if err != nil {
		return nil, fmt.Errorf("invalid YouTube URL: %w", err)
	}

	fmt.Printf("📼 Video ID: %s\n", videoID)

	// Use Innertube API
	playerResponse, err := c.fetchInnertubePlayerResponse(videoID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch player response: %w", err)
	}

	fmt.Println("✅ Successfully fetched data via Innertube API")

	video, err := ExtractVideoMetadata(playerResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to extract metadata: %w", err)
	}

	video.Formats, err = ExtractFormats(playerResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to extract formats: %w", err)
	}

	fmt.Printf("📦 Found %d formats\n", len(video.Formats))

	return &video, nil
}

// Download method
func (c *Client) Download(video *Video, outputPath string) error {
	if len(video.Formats) == 0 {
		return fmt.Errorf("no formats available")
	}

	format, err := SelectBestFormat(video.Formats)
	if err != nil {
		return err
	}

	fmt.Printf("📹 Selected: %s %s (itag %d)\n", format.QualityLabel, format.MimeType, format.Itag)

	if format.URL == "" {
		return fmt.Errorf("selected format requires signature deciphering (not implemented yet)")
	}

	fmt.Printf("⬇️  Downloading to: %s\n", outputPath)

	resp, err := c.httpClient.Get(format.URL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

// Innertube API call - using ANDROID client (more reliable)
func (c *Client) fetchInnertubePlayerResponse(videoID string) (map[string]interface{}, error) {
    apiURL := "https://www.youtube.com/youtubei/v1/player?key=AIzaSyAO_FJ2SlqU8Q4STEHLGCilw_Y9_11qcW8"

    payload := map[string]interface{}{
        "videoId": videoID,
        "context": map[string]interface{}{
            "client": map[string]interface{}{
                "clientName":    "ANDROID",
                "clientVersion": "19.45.36",
                "androidSdkVersion": 30,
            },
        },
        "contentCheckOk": true,
        "racyCheckOk":    true,
    }

    jsonPayload, _ := json.Marshal(payload)

    req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonPayload))
    if err != nil {
        return nil, err
    }

    req.Header.Set("User-Agent", "com.google.android.youtube/19.45.36 (Linux; U; Android 14) gzip")
    req.Header.Set("Accept", "application/json")
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Origin", "https://www.youtube.com")
    req.Header.Set("Referer", "https://www.youtube.com/watch?v="+videoID)

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("innertube API returned status: %d", resp.StatusCode)
    }

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }

    var result map[string]interface{}
    err = json.Unmarshal(body, &result)
    if err != nil {
        return nil, err
    }

    return result, nil
}