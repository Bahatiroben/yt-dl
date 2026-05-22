package youtube

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	_ "strings"
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

// fetchWatchPage with gzip support
func (c *Client) fetchWatchPage(videoID string) (string, error) {
	url := "https://www.youtube.com/watch?v=" + videoID

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Referer", "https://www.youtube.com/")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status code: %d", resp.StatusCode)
	}

	var reader io.Reader = resp.Body
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gzipReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return "", err
		}
		defer gzipReader.Close()
		reader = gzipReader
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
