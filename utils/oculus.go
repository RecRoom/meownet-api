package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

var oculusClient = &http.Client{Timeout: 10 * time.Second}

var ErrOculusValidationFailed = errors.New("oculus nonce validation failed")

var ErrOculusUserNotFound = errors.New("oculus user does not exist")

const oculusAppID = "1196298696890609"

type oculusNonceResponse struct {
	IsValid bool `json:"is_valid"`
	Error   *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    int    `json:"code"`
	} `json:"error"`
}

var oculusTransientCodes = map[int]bool{
	1: true,
	2: true,
}

const oculusMaxAttempts = 3

func ValidateOculusNonce(nonce, userID, appID string) error {
	if nonce == "" || userID == "" || appID == "" {
		return fmt.Errorf("%w: missing nonce, user_id, or app_id", ErrOculusValidationFailed)
	}
	secret := os.Getenv("OCULUS_APP_SECRET")
	if secret == "" {
		return fmt.Errorf("OCULUS_APP_SECRET not configured")
	}

	form := url.Values{}
	form.Set("nonce", nonce)
	form.Set("user_id", userID)
	form.Set("access_token", fmt.Sprintf("OC|%s|%s", appID, secret))

	var lastErr error
	for attempt := 1; attempt <= oculusMaxAttempts; attempt++ {
		retryable, err := validateOculusNonceOnce(form)
		if err == nil {
			return nil
		}
		lastErr = err
		if !retryable || attempt == oculusMaxAttempts {
			break
		}
		backoff := time.Duration(attempt*attempt) * 250 * time.Millisecond
		log.Printf("[OCULUS] transient validation error (attempt %d/%d), retrying in %v: %v", attempt, oculusMaxAttempts, backoff, err)
		time.Sleep(backoff)
	}
	return lastErr
}

type oculusUserNode struct {
	ID    string `json:"id"`
	Alias string `json:"alias"`
	Error *struct {
		Message     string `json:"message"`
		Type        string `json:"type"`
		Code        int    `json:"code"`
		IsTransient bool   `json:"is_transient"`
	} `json:"error"`
}

func ValidateOculusUserExists(userID string) error {
	if userID == "" {
		return fmt.Errorf("%w: missing user_id", ErrOculusUserNotFound)
	}
	if _, err := strconv.ParseUint(userID, 10, 64); err != nil {
		return fmt.Errorf("%w: non-numeric user_id", ErrOculusUserNotFound)
	}
	secret := os.Getenv("OCULUS_APP_SECRET")
	if secret == "" {
		return fmt.Errorf("OCULUS_APP_SECRET not configured")
	}

	reqURL := fmt.Sprintf(
		"https://graph.oculus.com/%s?fields=id&access_token=OC|%s|%s",
		userID, oculusAppID, secret,
	)

	var lastErr error
	for attempt := 1; attempt <= oculusMaxAttempts; attempt++ {
		retryable, err := validateOculusUserExistsOnce(reqURL, userID)
		if err == nil {
			return nil
		}
		lastErr = err
		if !retryable || attempt == oculusMaxAttempts {
			break
		}
		backoff := time.Duration(attempt*attempt) * 250 * time.Millisecond
		log.Printf("[OCULUS] transient user-lookup error (attempt %d/%d), retrying in %v: %v", attempt, oculusMaxAttempts, backoff, err)
		time.Sleep(backoff)
	}
	return lastErr
}

func validateOculusUserExistsOnce(reqURL, userID string) (retryable bool, err error) {
	resp, err := oculusClient.Get(reqURL)
	if err != nil {
		return true, err
	}
	defer resp.Body.Close()

	var node oculusUserNode
	if err := json.NewDecoder(resp.Body).Decode(&node); err != nil {
		return true, err
	}
	if node.Error != nil {
		retry := node.Error.IsTransient || oculusTransientCodes[node.Error.Code]
		return retry, fmt.Errorf("%w: oculus error %d: %s", ErrOculusUserNotFound, node.Error.Code, node.Error.Message)
	}
	if node.ID != userID {
		return false, fmt.Errorf("%w: id mismatch got=%s want=%s", ErrOculusUserNotFound, node.ID, userID)
	}
	return false, nil
}

func validateOculusNonceOnce(form url.Values) (retryable bool, err error) {
	req, err := http.NewRequest("POST", "https://graph.oculus.com/user_nonce_validate", strings.NewReader(form.Encode()))
	if err != nil {
		return false, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := oculusClient.Do(req)
	if err != nil {
		return true, err
	}
	defer resp.Body.Close()

	var result oculusNonceResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return true, err
	}
	if result.Error != nil {
		return oculusTransientCodes[result.Error.Code],
			fmt.Errorf("%w: oculus error %d: %s", ErrOculusValidationFailed, result.Error.Code, result.Error.Message)
	}
	if !result.IsValid {
		return false, fmt.Errorf("%w: nonce rejected", ErrOculusValidationFailed)
	}
	return false, nil
}
