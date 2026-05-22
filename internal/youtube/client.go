package youtube

import (
    "fmt"
    "io"
    "net/http"
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
    videoID, err := ExtractVideoID(url)
    if err != nil {
        return nil, fmt.Errorf("invalid YouTube URL: %w", err)
    }

    fmt.Printf("📼 Video ID: %s\n", videoID)

    // Fetch watch page
    pageHTML, err := c.fetchWatchPage(videoID)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch video page: %w", err)
    }

    fmt.Printf("📄 Fetched watch page (%d bytes)\n", len(pageHTML))

    // Extract player response JSON
    playerResponse, err := ExtractPlayerResponse(pageHTML)
    if err != nil {
        return nil, fmt.Errorf("failed to extract player response: %w", err)
    }

    fmt.Println("✅ Extracted ytInitialPlayerResponse")

    // Extract metadata
    video, err := ExtractVideoMetadata(playerResponse)
    if err != nil {
        return nil, fmt.Errorf("failed to extract metadata: %w", err)
    }

    fmt.Printf("🎥 Title : %s\n", video.Title)
    fmt.Printf("👤 Author: %s\n", video.Author)
    fmt.Printf("⏱️  Length: %d seconds\n", video.Length)

    return &video, nil
}

func (c *Client) fetchWatchPage(videoID string) (string, error) {
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