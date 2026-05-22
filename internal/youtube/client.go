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

	// Try Innertube API first
	playerResponse, err := c.fetchInnertubePlayerResponse(videoID)
	if err != nil {
		fmt.Printf("⚠️  Innertube API failed: %v\n", err)
		fmt.Println("🔄 Trying HTML page extraction...")
		// Fallback to HTML extraction
		playerResponse, err = c.fetchPlayerResponseFromHTML(videoID)
		if err != nil {
			return nil, fmt.Errorf("both API and HTML extraction failed: %w", err)
		}
	}

	fmt.Println("✅ Successfully fetched data")

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

// Innertube API call - using WEB client
func (c *Client) fetchInnertubePlayerResponse(videoID string) (map[string]interface{}, error) {
	apiURL := "https://www.youtube.com/youtubei/v1/player?key=AIzaSyAO_FJ2SlqU8Q4STEHLGCilw_Y9_11qcW8"

	payload := map[string]interface{}{
		"videoId": videoID,
		"context": map[string]interface{}{
			"client": map[string]interface{}{
				"clientName":    "WEB",
				"clientVersion": "2.20240522.00.00",
				"hl":            "en",
				"gl":            "US",
			},
		},
	}

	jsonPayload, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://www.youtube.com")
	req.Header.Set("Referer", "https://www.youtube.com/watch?v="+videoID)
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("❌ API Error (Status %d):\n%s\n", resp.StatusCode, string(body))
		return nil, fmt.Errorf("innertube API returned status: %d - %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	// Debug: Save full response to file
	prettyJSON, _ := json.MarshalIndent(result, "", "  ")
	os.WriteFile("innertube_response.json", prettyJSON, 0644)

	// Check playability status before accepting response
	if playability, ok := result["playabilityStatus"].(map[string]interface{}); ok {
		if status, ok := playability["status"].(string); ok && status != "OK" {
			if reason, ok := playability["reason"].(string); ok {
				fmt.Printf("❌ Playability check: %s - %s\n", status, reason)
			}
			return nil, fmt.Errorf("API returned non-OK status: %s", status)
		}
	}

	fmt.Println("💾 Full Innertube response saved to innertube_response.json")

	return result, nil
}

// fetchPlayerResponseFromHTML fetches the YouTube page and extracts ytInitialPlayerResponse
func (c *Client) fetchPlayerResponseFromHTML(videoID string) (map[string]interface{}, error) {
	url := fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	htmlContent := string(body)
	fmt.Println("📄 Fetched YouTube page, attempting to extract playerResponse...")

	playerResponse, err := ExtractPlayerResponse(htmlContent)
	if err != nil {
		return nil, fmt.Errorf("failed to extract from HTML: %w", err)
	}

	return playerResponse, nil
}
