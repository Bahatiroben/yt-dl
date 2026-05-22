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

    // Fetch the watch page
    pageHTML, err := c.fetchWatchPage(videoID)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch video page: %w", err)
    }

    fmt.Printf("📄 Successfully fetched watch page (%d bytes)\n", len(pageHTML))

    // Still placeholder response for now
    return &Video{
        ID:     videoID,
        Title:  "Placeholder Title",
        Author: "Placeholder Author",
        Length: 180,
    }, nil
}

// fetchWatchPage downloads the YouTube watch page HTML
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