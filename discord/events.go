package discord

import (
	"fmt"
	"time"
)

type InstanceClosedInfo struct {
	InstanceID int64
	RoomName   string
	Result     int
}

func SendInstanceClosed(info InstanceClosedInfo) {
	embed := Embed{
		Author:      &EmbedAuthor{Name: brandUsername},
		Title:       "Instance Closed",
		Description: buildInstanceClosedDesc(info),
		Color:       embedColor,
		Timestamp:   FormatTimestamp(time.Now()),
	}
	SendAsync(WebhookEvents, brandUsername, embed)
}

func buildInstanceClosedDesc(info InstanceClosedInfo) string {
	room := info.RoomName
	if room == "" {
		room = "a room"
	} else {
		room = "^" + room
	}
	return fmt.Sprintf("Instance #%d (%s) closed to new joiners after reported join result %d", info.InstanceID, room, info.Result)
}
