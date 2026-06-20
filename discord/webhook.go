package discord

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"
)

var httpClient = &http.Client{Timeout: 10 * time.Second}

// each webhook link
const (
	WebhookImages = "https://discord.com/api/webhooks/1492554174867570719/dFb0DNCorZKExXlQ7QuGU_Nx02TtY9gb2buslbEi5lte08tuZvfU9aLzfWScTuM068LV"
	WebhookEvents = "https://discord.com/api/webhooks/1498469401211568128/kCMQZaiSlz1FLY7WRuhdk_Uqot5Q3T8Aav7yXQ9y0usbAG866ysjXZtbmcV7eLCK7YMX"
)

type Embed struct {
	Title       string       `json:"title,omitempty"`
	Description string       `json:"description,omitempty"`
	URL         string       `json:"url,omitempty"`
	Color       int          `json:"color,omitempty"`
	Timestamp   string       `json:"timestamp,omitempty"`
	Author      *EmbedAuthor `json:"author,omitempty"`
	Footer      *EmbedFooter `json:"footer,omitempty"`
	Image       *EmbedImage  `json:"image,omitempty"`
	Thumbnail   *EmbedImage  `json:"thumbnail,omitempty"`
	Fields      []EmbedField `json:"fields,omitempty"`
}

type EmbedAuthor struct {
	Name    string `json:"name,omitempty"`
	URL     string `json:"url,omitempty"`
	IconURL string `json:"icon_url,omitempty"`
}

type EmbedFooter struct {
	Text    string `json:"text,omitempty"`
	IconURL string `json:"icon_url,omitempty"`
}

type EmbedImage struct {
	URL string `json:"url,omitempty"`
}

type EmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

type webhookPayload struct {
	Username    string   `json:"username,omitempty"`
	AvatarURL   string   `json:"avatar_url,omitempty"`
	Content     string   `json:"content,omitempty"`
	Embeds      []Embed  `json:"embeds,omitempty"`
	ThreadName  string   `json:"thread_name,omitempty"`
	AppliedTags []string `json:"applied_tags,omitempty"`
}

func Send(webhookURL, username string, embeds ...Embed) {
	send(webhookURL, webhookPayload{Username: username, Embeds: embeds})
}

func SendForum(webhookURL, username, threadName string, tags []string, embeds ...Embed) {
	send(webhookURL, webhookPayload{
		Username:    username,
		ThreadName:  threadName,
		AppliedTags: tags,
		Embeds:      embeds,
	})
}

func send(webhookURL string, payload webhookPayload) {
	body, err := json.Marshal(payload)
	if err != nil {
		log.Printf("[DISCORD] marshal error: %v", err)
		return
	}

	req, err := http.NewRequest(http.MethodPost, webhookURL, bytes.NewReader(body))
	if err != nil {
		log.Printf("[DISCORD] request build error: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Printf("[DISCORD] send error: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		log.Printf("[DISCORD] send failed status=%d body=%s", resp.StatusCode, string(msg))
		return
	}
	log.Printf("[DISCORD] sent embed status=%d", resp.StatusCode)
}

// send async so caller doesn't block on discord
func SendAsync(webhookURL, username string, embeds ...Embed) {
	go Send(webhookURL, username, embeds...)
}

func SendForumAsync(webhookURL, username, threadName string, tags []string, embeds ...Embed) {
	go SendForum(webhookURL, username, threadName, tags, embeds...)
}

func FormatTimestamp(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}
