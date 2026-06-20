package utils

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	openAIModerationEndpoint = "https://api.openai.com/v1/moderations"
	openAIModerationModel    = "omni-moderation-latest"

	openRouterEndpoint     = "https://openrouter.ai/api/v1/chat/completions"
	openRouterDefaultModel = "google/gemini-3.1-flash-lite-20260507"

	moderationTimeout = 5 * time.Second

	moderationSystemPrompt = `You are a strict content moderation classifier. ` +
		`Decide whether the user's message contains content that is sexual, hateful or ` +
		`discriminatory, harassment, threats or violence, self-harm, or otherwise unsafe. ` +
		`Respond with ONLY a compact JSON object: {"flagged": true} or {"flagged": false}. ` +
		`Output no other text.`
)

var moderationClient = &http.Client{Timeout: moderationTimeout}

func IsTextFlagged(text string) bool {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return false
	}

	if IsBlocklisted(trimmed) {
		return true
	}

	if key := os.Getenv("OPENROUTER_API_KEY"); key != "" {
		return openRouterFlagged(trimmed, key)
	}
	if key := os.Getenv("OPENAI_API_KEY"); key != "" {
		return openAIFlagged(trimmed, key)
	}
	return false
}

func IsAnyTextFlagged(texts ...string) bool {
	for _, t := range texts {
		if IsTextFlagged(t) {
			return true
		}
	}
	return false
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model          string          `json:"model"`
	Messages       []chatMessage   `json:"messages"`
	Temperature    float64         `json:"temperature"`
	MaxTokens      int             `json:"max_tokens"`
	ResponseFormat *responseFormat `json:"response_format,omitempty"`
}

type responseFormat struct {
	Type string `json:"type"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func openRouterFlagged(text, apiKey string) bool {
	model := os.Getenv("OPENROUTER_MODEL")
	if model == "" {
		model = openRouterDefaultModel
	}

	body, err := json.Marshal(chatRequest{
		Model:       model,
		Temperature: 0,
		MaxTokens:   16,
		Messages: []chatMessage{
			{Role: "system", Content: moderationSystemPrompt},
			{Role: "user", Content: text},
		},
		ResponseFormat: &responseFormat{Type: "json_object"},
	})
	if err != nil {
		log.Printf("[MODERATION] marshal error: %v", err)
		return false
	}

	req, err := http.NewRequest(http.MethodPost, openRouterEndpoint, bytes.NewReader(body))
	if err != nil {
		log.Printf("[MODERATION] request error: %v", err)
		return false
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := moderationClient.Do(req)
	if err != nil {
		log.Printf("[MODERATION] http error: %v", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("[MODERATION] openrouter non-200 status: %d", resp.StatusCode)
		return false
	}

	var decoded chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		log.Printf("[MODERATION] decode error: %v", err)
		return false
	}
	if len(decoded.Choices) == 0 {
		return false
	}

	content := strings.TrimSpace(decoded.Choices[0].Message.Content)
	var parsed struct {
		Flagged bool `json:"flagged"`
	}
	if err := json.Unmarshal([]byte(content), &parsed); err == nil {
		if parsed.Flagged {
			log.Printf("[MODERATION] flagged text: %q", text)
		}
		return parsed.Flagged
	}

	// Fallback if the model wrapped the JSON in prose/code fences.
	flagged := strings.Contains(strings.ToLower(content), `"flagged": true`) ||
		strings.Contains(strings.ToLower(content), `"flagged":true`)
	if flagged {
		log.Printf("[MODERATION] flagged text: %q", text)
	}
	return flagged
}

type moderationRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

type moderationResult struct {
	Flagged bool `json:"flagged"`
}

type moderationResponse struct {
	Results []moderationResult `json:"results"`
}

func openAIFlagged(text, apiKey string) bool {
	body, err := json.Marshal(moderationRequest{
		Model: openAIModerationModel,
		Input: text,
	})
	if err != nil {
		log.Printf("[MODERATION] marshal error: %v", err)
		return false
	}

	req, err := http.NewRequest(http.MethodPost, openAIModerationEndpoint, bytes.NewReader(body))
	if err != nil {
		log.Printf("[MODERATION] request error: %v", err)
		return false
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := moderationClient.Do(req)
	if err != nil {
		log.Printf("[MODERATION] http error: %v", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("[MODERATION] non-200 status: %d", resp.StatusCode)
		return false
	}

	var decoded moderationResponse
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		log.Printf("[MODERATION] decode error: %v", err)
		return false
	}

	for _, r := range decoded.Results {
		if r.Flagged {
			log.Printf("[MODERATION] flagged text: %q", text)
			return true
		}
	}
	return false
}
