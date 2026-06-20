package discord

import (
	"fmt"
	"time"
)

const WebhookAnticheat = "https://discord.com/api/webhooks/1512540469362430025/gzX6J7fiMeT5un60TsJu5AC6kPMBIWsiGuG3H0ZZL8Wcqe2v8X9IzZZplOTLuuIEi9ru"

type AnticheatInfo struct {
	DetectionType string
	DeviceID      string
	Details       string
	IP            string
	AccountID     uint
	AccountName   string
}

func SendAnticheatDetection(info AnticheatInfo) {
	embed := Embed{
		Title:     "Anticheat Detection",
		Timestamp: FormatTimestamp(time.Now()),
		Fields: []EmbedField{
			{Name: "Type", Value: info.DetectionType, Inline: true},
			{Name: "Account", Value: formatAccountRef(info.AccountName, info.AccountID), Inline: true},
			{Name: "Device", Value: fmt.Sprintf("`%s`", info.DeviceID)},
			{Name: "Details", Value: truncateDetails(info.Details)},
			{Name: "IP", Value: ipOrUnknown(info.IP), Inline: true},
		},
	}
	SendAsync(WebhookAnticheat, brandUsername, embed)
}

func formatAccountRef(name string, id uint) string {
	if id == 0 {
		return "_(unlinked)_"
	}
	if name == "" {
		return fmt.Sprintf("#%d", id)
	}
	return fmt.Sprintf("@%s (#%d)", name, id)
}

func ipOrUnknown(ip string) string {
	if ip == "" {
		return "_(unknown)_"
	}
	return fmt.Sprintf("`%s`", ip)
}
