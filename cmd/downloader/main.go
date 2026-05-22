package main

import (
	"fmt"
	"log"

	"github.com/Bahatiroben/yt-dl/internal/utils"
	"github.com/Bahatiroben/yt-dl/internal/youtube"
)

func main() {
	fmt.Println("🚀 My YouTube Downloader")

	videoURL := "https://www.youtube.com/watch?v=5XAUYY8kdcg&list=RD5XAUYY8kdcg&start_radio=1"

	client := youtube.NewClient()

	video, err := client.GetVideo(videoURL)
	if err != nil {
		log.Fatalf("❌ Failed: %v", err)
	}

	// Create filename
	filename := utils.SanitizeFilename(video.Title) + ".mp4"

	err = client.Download(video, filename)
	if err != nil {
		log.Fatalf("❌ Download failed: %v", err)
	}

	fmt.Println("✅ Download completed successfully!")
}
