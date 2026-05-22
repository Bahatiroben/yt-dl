package utils

import "strings"

func SanitizeFilename(name string) string {
    invalid := []rune{'/', '\\', ':', '*', '?', '"', '<', '>', '|'}
    for _, r := range invalid {
        name = strings.ReplaceAll(name, string(r), "_")
    }
    return name
}