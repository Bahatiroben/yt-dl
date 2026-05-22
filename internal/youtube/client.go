package youtube

import (
    "fmt"
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

    fmt.Printf("📼 Extracted Video ID: %s\n", videoID)

    return &Video{
        ID:     videoID,
        Title:  "Placeholder Title",
        Author: "Placeholder Author",
        Length: 180,
    }, nil
}