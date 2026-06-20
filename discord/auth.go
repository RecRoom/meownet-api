package discord

import (
	"fmt"
	"time"
)

const WebhookAuth = "https://discord.com/api/webhooks/1511898037700657263/sGa_xRm-Kn3OpDQAmk07nP-r1tFlWky3wA6E-t9QNKN6YRFDCz07jNsbu2zIQv0nPPuz"

type AuthErrorInfo struct {
	Stage      string
	Platform   string
	PlatformID string
	Username   string
	AccountID  uint
	IP         string
	Reason     string
}

func SendAuthError(info AuthErrorInfo) {
	embed := Embed{
		Title:     "Auth Error",
		Timestamp: FormatTimestamp(time.Now()),
		Fields: []EmbedField{
			{Name: "Stage", Value: stageOrUnknown(info.Stage), Inline: true},
			{Name: "Platform", Value: platformRef(info.Platform, info.PlatformID), Inline: true},
			{Name: "Account", Value: authAccountRef(info.Username, info.AccountID), Inline: true},
			{Name: "Reason", Value: truncateDetails(info.Reason)},
			{Name: "IP", Value: ipOrUnknown(info.IP), Inline: true},
		},
	}
	SendAsync(WebhookAuth, brandUsername, embed)
}

func authAccountRef(username string, id uint) string {
	switch {
	case id != 0 && username != "":
		return fmt.Sprintf("@%s (#%d)", username, id)
	case id != 0:
		return fmt.Sprintf("#%d", id)
	case username != "":
		return fmt.Sprintf("@%s", username)
	default:
		return "none"
	}
}

func stageOrUnknown(s string) string {
	if s == "" {
		return "unknown"
	}
	return s
}

func platformRef(platform, platformID string) string {
	if platform == "" && platformID == "" {
		return "unknown"
	}
	if platformID == "" {
		return platform
	}
	if platform == "" {
		return fmt.Sprintf("`%s`", platformID)
	}
	return fmt.Sprintf("%s `%s`", platform, platformID)
}
