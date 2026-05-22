package youtube

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	_ "github.com/Bahatiroben/yt-dl/internal/utils"
)

type Client struct {
	httpClient *http.Client
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{},
	}
}

func (c *Client) GetVideo(url string) (*Video, error) {
	// ... (keep all previous code until format extraction)

	videoID, err := ExtractVideoID(url)
	if err != nil {
		return nil, fmt.Errorf("invalid YouTube URL: %w", err)
	}

	fmt.Printf("📼 Video ID: %s\n", videoID)

	pageHTML, err := c.fetchWatchPage(videoID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch video page: %w", err)
	}

	fmt.Printf("📄 Fetched watch page (%d bytes)\n", len(pageHTML))

	playerResponse, err := ExtractPlayerResponse(pageHTML)
	if err != nil {
		return nil, fmt.Errorf("failed to extract player response: %w", err)
	}

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

// Download downloads a specific format to disk
func (c *Client) Download(video *Video, outputPath string) error {
	if len(video.Formats) == 0 {
		return fmt.Errorf("no formats available")
	}

	format, err := SelectBestFormat(video.Formats)
	if err != nil {
		return err
	}

	fmt.Printf("📹 Selected: %s %s (itag %d)\n", format.QualityLabel, format.MimeType, format.Itag)

	downloadURL := format.URL

	// If no direct URL, we need deciphering
	if downloadURL == "" {
		fmt.Println("🔐 Format requires signature deciphering...")
		// TODO: Implement deciphering in next steps
		return fmt.Errorf("deciphering required but not implemented yet")
	}

	fmt.Printf("⬇️  Downloading to: %s\n", outputPath)

	resp, err := c.httpClient.Get(downloadURL)
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

func (c *Client) fetchWatchPage(videoID string) (string, error) {
	// ... (keep existing fetchWatchPage function)
	url := "https://www.youtube.com/watch?v=" + videoID

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// Add this method to Client
func (c *Client) getDecipheredURL(format Format, html string) (string, error) {
	if format.URL != "" {
		return format.URL, nil // Already has direct URL
	}

	// In real implementation, we would need to:
	// 1. Get player JS
	// 2. Extract decipher function
	// 3. Apply transformations to signature

	// For now, we'll return a helpful message
	return "", errors.New("signature deciphering not implemented yet")
}
