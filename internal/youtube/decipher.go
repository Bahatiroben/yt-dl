package youtube

import (
    "errors"
    _ "fmt"
    "regexp"
    "strings"
)

// Decipherer handles YouTube signature decryption
type Decipherer struct {
    operations []string
}

func NewDecipherer() *Decipherer {
    return &Decipherer{}
}

// ExtractPlayerJSURL finds the URL of the player JavaScript
func ExtractPlayerJSURL(html string) (string, error) {
    // Multiple possible patterns for player JS
    patterns := []string{
        `player_[\w\d]+\.js`,
        `https?://[^"]+player_[\w\d/]+\.js`,
    }

    for _, pattern := range patterns {
        re := regexp.MustCompile(pattern)
        match := re.FindString(html)
        if match != "" {
            if !strings.HasPrefix(match, "http") {
                match = "https://www.youtube.com" + match
            }
            return match, nil
        }
    }

    return "", errors.New("could not find player JavaScript URL")
}