package main

import (
	"fmt"
	"log"

	"github.com/Bahatiroben/yt-dl/internal/utils"
	"github.com/Bahatiroben/yt-dl/internal/youtube"
)

func main() {
    fmt.Println("🚀 My YouTube Downloader")

    // You can change this URL
    videoURL := "https://www.youtube.com/watch?v=borrrGqCbd0&list=RD5XAUYY8kdcg&index=3"

    client := youtube.NewClient()

    video, err := client.GetVideo(videoURL)
    if err != nil {
        log.Fatalf("❌ Failed to get video: %v", err)
    }

    fmt.Printf("\n🎥 Title : %s\n", video.Title)
    fmt.Printf("👤 Author: %s\n", video.Author)
    fmt.Printf("📦 Total formats: %d\n\n", len(video.Formats))

    // Create safe filename
    filename := utils.SanitizeFilename(video.Title) + ".mp4"

    err = client.Download(video, filename)
    if err != nil {
        log.Fatalf("❌ Download failed: %v", err)
    }

    fmt.Printf("✅ Successfully downloaded: %s\n", filename)
}