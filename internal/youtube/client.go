package youtube

import (
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
    // TODO: Implement actual logic to fetch video details from YouTube
    return &Video{
        Title:  "Placeholder Title",
        Author: "Placeholder Author",
        Length: 180,
    }, nil
}