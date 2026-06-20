package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

var steamClient = &http.Client{Timeout: 10 * time.Second}

const minSteamAccountAgeDays = 14

var ErrSteamStandingFailed = errors.New("steam account standing check failed")

type steamAuthResponse struct {
	Response struct {
		Params *struct {
			Result  string `json:"result"`
			SteamID string `json:"steamid"`
		} `json:"params"`
		Error *struct {
			ErrorCode int    `json:"errorcode"`
			ErrorDesc string `json:"errordesc"`
		} `json:"error"`
	} `json:"response"`
}

type steamPlayerSummariesResponse struct {
	Response struct {
		Players []struct {
			SteamID                  string `json:"steamid"`
			CommunityVisibilityState int    `json:"communityvisibilitystate"`
			ProfileState             int    `json:"profilestate"`
			TimeCreated              int64  `json:"timecreated"`
		} `json:"players"`
	} `json:"response"`
}

type steamPlayerBansResponse struct {
	Players []struct {
		SteamID         string `json:"SteamId"`
		CommunityBanned bool   `json:"CommunityBanned"`
	} `json:"players"`
}

// verify steam ticket, return steamid
func ValidateSteamTicket(ticket, appID string) (string, error) {
	apiKey := os.Getenv("STEAM_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("STEAM_API_KEY not configured")
	}

	var lastErr error
	for _, candidate := range candidateAppIDs(appID) {
		steamID, err := authenticateSteamTicket(apiKey, candidate, ticket)
		if err == nil {
			return steamID, nil
		}
		if !errors.Is(err, errSteamTicketOtherApp) {
			return "", err
		}
		lastErr = err
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("no steam appid configured to verify ticket")
	}
	return "", lastErr
}

var errSteamTicketOtherApp = errors.New("steam ticket for other app")

var steamAllowedAppIDs = []string{"480", "471710"}

func candidateAppIDs(appID string) []string {
	var out []string
	seen := map[string]bool{}
	add := func(id string) {
		id = strings.TrimSpace(id)
		if id == "" || seen[id] {
			return
		}
		seen[id] = true
		out = append(out, id)
	}
	add(appID)
	for _, id := range steamAllowedAppIDs {
		add(id)
	}
	return out
}

func authenticateSteamTicket(apiKey, appID, ticket string) (string, error) {
	url := fmt.Sprintf(
		"https://api.steampowered.com/ISteamUserAuth/AuthenticateUserTicket/v1/?key=%s&appid=%s&ticket=%s",
		apiKey, appID, ticket,
	)
	resp, err := steamClient.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result steamAuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if result.Response.Error != nil {
		code := result.Response.Error.ErrorCode
		if code == 102 {
			return "", fmt.Errorf("%w: appid=%s: %s", errSteamTicketOtherApp, appID, result.Response.Error.ErrorDesc)
		}
		return "", fmt.Errorf("steam error %d: %s", code, result.Response.Error.ErrorDesc)
	}
	if result.Response.Params == nil || result.Response.Params.Result != "OK" {
		return "", fmt.Errorf("steam auth failed")
	}
	return result.Response.Params.SteamID, nil
}

func ValidateSteamAccountStanding(steamID string) error {
	apiKey := os.Getenv("STEAM_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("STEAM_API_KEY not configured")
	}

	summaryURL := fmt.Sprintf(
		"https://api.steampowered.com/ISteamUser/GetPlayerSummaries/v2/?key=%s&steamids=%s",
		apiKey, steamID,
	)
	resp, err := steamClient.Get(summaryURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var summary steamPlayerSummariesResponse
	if err := json.NewDecoder(resp.Body).Decode(&summary); err != nil {
		return err
	}
	if len(summary.Response.Players) == 0 {
		return fmt.Errorf("%w: Steam profile not found", ErrSteamStandingFailed)
	}
	player := summary.Response.Players[0]

	if player.TimeCreated != 0 {
		created := time.Unix(player.TimeCreated, 0)
		ageDays := int(time.Since(created).Hours() / 24)
		if ageDays < minSteamAccountAgeDays {
			return fmt.Errorf("%w: your Steam account must be at least %d days old (currently %d days)", ErrSteamStandingFailed, minSteamAccountAgeDays, ageDays)
		}
	}
	if player.ProfileState != 1 {
		return fmt.Errorf("%w: your Steam profile must be set up before you can create an account", ErrSteamStandingFailed)
	}

	bansURL := fmt.Sprintf(
		"https://api.steampowered.com/ISteamUser/GetPlayerBans/v1/?key=%s&steamids=%s",
		apiKey, steamID,
	)
	bresp, err := steamClient.Get(bansURL)
	if err != nil {
		return err
	}
	defer bresp.Body.Close()

	var bans steamPlayerBansResponse
	if err := json.NewDecoder(bresp.Body).Decode(&bans); err != nil {
		return err
	}
	if len(bans.Players) == 0 {
		return fmt.Errorf("%w: unable to verify Steam ban status", ErrSteamStandingFailed)
	}
	b := bans.Players[0]
	if b.CommunityBanned {
		return fmt.Errorf("%w: Steam accounts with active community bans are not allowed", ErrSteamStandingFailed)
	}

	return nil
}
