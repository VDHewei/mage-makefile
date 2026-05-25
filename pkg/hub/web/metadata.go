package web

import (
	"strings"
	"time"
)

// parseMagefileMetadata 从字符串中解析 magefile.go 元数据。
func parseMagefileMetadata(code string) Metadata {
	lines := strings.Split(code, "\n")
	meta := Metadata{
		Name:        "unknown",
		Description: "magefile snippet",
		Tags:        []string{},
		Author:      "unknown",
		Platform:    "",
		Engine:      "",
		Metadata:    make(map[string]string),
		CreatedAt:   time.Now(),
	}

	for _, line := range lines {
		// Look for // @hub lines
		if strings.HasPrefix(line, "// @hub") {
			parts := strings.Split(line[7:], " ")
			for _, part := range parts {
				part = strings.TrimSpace(part)
				if part == "" {
					continue
				}
				if strings.HasPrefix(part, "name=") {
					meta.Name = strings.Trim(strings.TrimPrefix(part, "name="), "'\"")
				} else if strings.HasPrefix(part, "description=") {
					meta.Description = strings.Trim(strings.TrimPrefix(part, "description="), "'\"")
				} else if strings.HasPrefix(part, "author=") {
					meta.Author = strings.Trim(strings.TrimPrefix(part, "author="), "'\"")
				} else if strings.HasPrefix(part, "platform=") {
					meta.Platform = strings.Trim(strings.TrimPrefix(part, "platform="), "'\"")
				} else if strings.HasPrefix(part, "engine=") {
					meta.Engine = strings.Trim(strings.TrimPrefix(part, "engine="), "'\"")
				} else if strings.HasPrefix(part, "tags=") {
					tagsStr := strings.Trim(strings.TrimPrefix(part, "tags="), "'\"")
					meta.Tags = strings.Split(tagsStr, ",")
				}
			}
		}
	}

	// If no metadata found, use filename
	// (filename-based fallback not applicable for string input)
	return meta
}
