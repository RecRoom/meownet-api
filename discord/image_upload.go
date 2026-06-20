package discord

import (
	"fmt"
	"strings"
	"time"
)

const (
	brandUsername = "meow.net"
	embedColor    = 0x808080
)

type ImageUploadInfo struct {
	ImageURL    string
	Uploader    string
	PlayerNames []string
	RoomName    string
	RoomOwner   string
}

// SendImageUploaded posts an "Image Uploaded" embed
func SendImageUploaded(info ImageUploadInfo) {
	embed := Embed{
		Author: &EmbedAuthor{
			Name: brandUsername,
		},
		Title:       "Image Uploaded",
		Description: buildDescription(info),
		Color:       embedColor,
		Timestamp:   FormatTimestamp(time.Now()),
	}

	if info.ImageURL != "" {
		embed.Image = &EmbedImage{URL: info.ImageURL}
	}
	SendAsync(WebhookImages, brandUsername, embed)
}

func buildDescription(info ImageUploadInfo) string {
	var b strings.Builder

	if info.Uploader != "" {
		b.WriteString(fmt.Sprintf("Picture taken by @%s", info.Uploader))
	} else {
		b.WriteString("Picture taken")
	}

	mentions := mentionOtherPlayers(info)
	if mentions != "" {
		b.WriteString(" with ")
		b.WriteString(mentions)
	}

	if info.RoomName != "" {
		b.WriteString(" in ")
		if info.RoomOwner != "" && info.RoomOwner != info.Uploader {
			b.WriteString(fmt.Sprintf(info.RoomName))
		} else {
			b.WriteString(fmt.Sprintf("^%s", info.RoomName))
		}
	}
	return b.String()
}

func mentionOtherPlayers(info ImageUploadInfo) string {
	if len(info.PlayerNames) == 0 {
		return ""
	}
	seen := map[string]bool{}
	if info.Uploader != "" {
		seen[info.Uploader] = true
	}
	var out []string

	for _, n := range info.PlayerNames {
		n = strings.TrimSpace(n)
		if n == "" || seen[n] {
			continue
		}
		seen[n] = true
		out = append(out, "@"+n)
	}
	return strings.Join(out, ", ")
}
