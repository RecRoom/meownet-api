package discord

import (
	"fmt"
	"strings"
	"time"
)

const WebhookReports = "https://discord.com/api/webhooks/1513343098392940616/Eq_ssI-txBKIbjI8t-2tiag9Apwy9qD-_GkHnuygJYL5db0gjsUUS-svmQlykRSfToMA"

const (
	tagPlayerReport    = "1513344601711186080"
	tagInventionReport = "1513344591619948605"
	tagClubReport      = "1513344556094062733"
	tagRoomReport      = "1513991045745217556"
)

func reportTags(ids ...string) []string {
	var tags []string
	for _, id := range ids {
		if id != "" {
			tags = append(tags, id)
		}
	}
	return tags
}

type PlayerReportInfo struct {
	ReporterID       uint
	ReporterName     string
	ReportedID       uint
	ReportedName     string
	CategoryID       int
	CategoryName     string
	Details          string
	HeightReporter   float64
	HeightReported   float64
	RoomID           int64
	RoomInstanceType int
}

func SendPlayerReport(info PlayerReportInfo) {
	embed := Embed{
		Author:    &EmbedAuthor{Name: brandUsername},
		Title:     "New Player Report",
		Color:     0xE74C3C,
		Timestamp: FormatTimestamp(time.Now()),
		Fields: []EmbedField{
			{Name: "Reporter", Value: formatPlayerRef(info.ReporterName, info.ReporterID), Inline: true},
			{Name: "Reported", Value: formatPlayerRef(info.ReportedName, info.ReportedID), Inline: true},
			{Name: "Category", Value: formatCategory(info.CategoryName, info.CategoryID), Inline: true},
			{Name: "Details", Value: truncateDetails(info.Details)},
			{Name: "Heights (m)", Value: fmt.Sprintf("reporter %.2f / reported %.2f", info.HeightReporter, info.HeightReported), Inline: true},
			{Name: "Room", Value: fmt.Sprintf("id %d (type %d)", info.RoomID, info.RoomInstanceType), Inline: true},
		},
	}
	threadName := truncateThreadName(fmt.Sprintf("Player: %s → %s", info.ReporterName, info.ReportedName))
	SendForumAsync(WebhookReports, brandUsername, threadName, reportTags(tagPlayerReport), embed)
}

type InventionReportInfo struct {
	ReporterID    uint
	ReporterName  string
	InventionID   int64
	InventionName string
	CreatorID     uint
	CreatorName   string
	CategoryID    int
	Details       string
}

func SendInventionReport(info InventionReportInfo) {
	embed := Embed{
		Title:     "New Invention Report",
		Timestamp: FormatTimestamp(time.Now()),
		Fields: []EmbedField{
			{Name: "Reporter", Value: formatPlayerRef(info.ReporterName, info.ReporterID), Inline: true},
			{Name: "Invention", Value: formatContentRef(info.InventionName, info.InventionID), Inline: true},
			{Name: "Creator", Value: formatPlayerRef(info.CreatorName, info.CreatorID), Inline: true},
			{Name: "Category", Value: formatCategory(contentReportCategoryName(info.CategoryID), info.CategoryID), Inline: true},
			{Name: "Details", Value: truncateDetails(info.Details)},
		},
	}
	threadName := truncateThreadName(fmt.Sprintf("Invention: %s", formatContentRef(info.InventionName, info.InventionID)))
	SendForumAsync(WebhookReports, brandUsername, threadName, reportTags(tagInventionReport), embed)
}

type ClubReportInfo struct {
	ReporterID   uint
	ReporterName string
	ClubID       int64
	ClubName     string
	CategoryID   int
	Details      string
}

func SendClubReport(info ClubReportInfo) {
	embed := Embed{
		Title:     "New Club Report",
		Timestamp: FormatTimestamp(time.Now()),
		Fields: []EmbedField{
			{Name: "Reporter", Value: formatPlayerRef(info.ReporterName, info.ReporterID), Inline: true},
			{Name: "Club", Value: formatContentRef(info.ClubName, info.ClubID), Inline: true},
			{Name: "Category", Value: formatCategory(contentReportCategoryName(info.CategoryID), info.CategoryID), Inline: true},
			{Name: "Details", Value: truncateDetails(info.Details)},
		},
	}
	threadName := truncateThreadName(fmt.Sprintf("Club: %s", formatContentRef(info.ClubName, info.ClubID)))
	SendForumAsync(WebhookReports, brandUsername, threadName, reportTags(tagClubReport), embed)
}

type RoomReportInfo struct {
	ReporterID   uint
	ReporterName string
	RoomID       int64
	RoomName     string
	CreatorID    uint
	CreatorName  string
	CategoryID   int
	Details      string
}

func SendRoomReport(info RoomReportInfo) {
	embed := Embed{
		Title:     "New Room Report",
		Timestamp: FormatTimestamp(time.Now()),
		Fields: []EmbedField{
			{Name: "Reporter", Value: formatPlayerRef(info.ReporterName, info.ReporterID), Inline: true},
			{Name: "Room", Value: formatContentRef(info.RoomName, info.RoomID), Inline: true},
			{Name: "Creator", Value: formatPlayerRef(info.CreatorName, info.CreatorID), Inline: true},
			{Name: "Category", Value: formatCategory(contentReportCategoryName(info.CategoryID), info.CategoryID), Inline: true},
			{Name: "Details", Value: truncateDetails(info.Details)},
		},
	}
	threadName := truncateThreadName(fmt.Sprintf("Room: %s", formatContentRef(info.RoomName, info.RoomID)))
	SendForumAsync(WebhookReports, brandUsername, threadName, reportTags(tagRoomReport), embed)
}

func contentReportCategoryName(c int) string {
	switch c {
	case -1:
		return "Unknown"
	case 0:
		return "CoC: Discriminatory"
	case 1:
		return "CoC: Sexual"
	case 2:
		return "CoC: Trolling"
	case 3:
		return "Misleading"
	case 4:
		return "Other"
	}
	return ""
}

func formatContentRef(name string, id int64) string {
	if name == "" {
		return fmt.Sprintf("#%d", id)
	}
	return fmt.Sprintf("%s (#%d)", name, id)
}

func formatPlayerRef(name string, id uint) string {
	if name == "" {
		return fmt.Sprintf("#%d", id)
	}
	return fmt.Sprintf("@%s (#%d)", name, id)
}

func formatCategory(name string, id int) string {
	if name == "" {
		return fmt.Sprintf("%d", id)
	}
	return fmt.Sprintf("%s (%d)", name, id)
}

func truncateThreadName(name string) string {
	name = strings.TrimSpace(name)
	if r := []rune(name); len(r) > 100 {
		return string(r[:99]) + "…"
	}
	return name
}

func truncateDetails(d string) string {
	d = strings.TrimSpace(d)
	if d == "" {
		return "_(none)_"
	}
	if len(d) > 1000 {
		return d[:1000] + "…"
	}
	return d
}
