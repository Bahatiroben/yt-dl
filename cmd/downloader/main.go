package main

import (
    "fmt"
    "log"

    "github.com/Bahatiroben/yt-dl/internal/youtube"
)

func main() {
    fmt.Println("🚀 My YouTube Downloader Starting...")

    // Hardcoded URL for now (we'll make it configurable later)
    videoURL := "https://www.youtube.com/watch?v=dQw4w9wgxcq"

    client := youtube.NewClient()

    video, err := client.GetVideo(videoURL)
    if err != nil {
        log.Fatalf("❌ Failed: %v", err)
    }

    fmt.Printf("✅ Success!\n")
    fmt.Printf("🎥 Title : %s\n", video.Title)
    fmt.Printf("👤 Author: %s\n", video.Author)
    fmt.Printf("📏 Length: %d seconds\n", video.Length)
    fmt.Printf("📦 Found %d formats\n", len(video.Formats))
}