package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

var proxyCheckClient = &http.Client{Timeout: 5 * time.Second}

const proxyCheckOKTTL = 24 * time.Hour

var (
	proxyCheckCache   = map[string]time.Time{}
	proxyCheckCacheMu sync.Mutex
)

func proxyCheckCacheGet(ip string) bool {
	proxyCheckCacheMu.Lock()
	defer proxyCheckCacheMu.Unlock()
	exp, ok := proxyCheckCache[ip]
	if !ok {
		return false
	}
	if time.Now().After(exp) {
		delete(proxyCheckCache, ip)
		return false
	}
	return true
}

func proxyCheckCachePut(ip string) {
	proxyCheckCacheMu.Lock()
	defer proxyCheckCacheMu.Unlock()
	proxyCheckCache[ip] = time.Now().Add(proxyCheckOKTTL)
}

func ClientIP(r *http.Request) string {
	if ip := r.Header.Get("CF-Connecting-IP"); ip != "" {
		return strings.TrimSpace(ip)
	}
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return strings.TrimSpace(ip)
	}
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		parts := strings.Split(ip, ",")
		return strings.TrimSpace(parts[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func IsProxy(ip string) (bool, string, error) {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return false, "", fmt.Errorf("invalid ip %q", ip)
	}
	if parsed.IsLoopback() || parsed.IsPrivate() || parsed.IsUnspecified() {
		return false, "", nil
	}
	if proxyCheckCacheGet(ip) {
		return false, "", nil
	}

	endpoint := fmt.Sprintf("https://proxycheck.io/v2/%s?vpn=1&asn=0", ip)
	if key := os.Getenv("PROXYCHECK_API_KEY"); key != "" {
		endpoint += "&key=" + key
	}

	resp, err := proxyCheckClient.Get(endpoint)
	if err != nil {
		return false, "", err
	}
	defer resp.Body.Close()

	var raw map[string]json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return false, "", err
	}

	entryRaw, ok := raw[ip]
	if !ok {
		proxyCheckCachePut(ip)
		return false, "", nil
	}
	var entry struct {
		Proxy string `json:"proxy"`
		Type  string `json:"type"`
	}
	if err := json.Unmarshal(entryRaw, &entry); err != nil {
		return false, "", err
	}
	isProxy := strings.EqualFold(entry.Proxy, "true")
	if !isProxy {
		proxyCheckCachePut(ip)
	}
	return isProxy, entry.Type, nil
}

func CheckProxy(r *http.Request) error {
	switch strings.ToLower(os.Getenv("PROXYCHECK_DISABLED")) {
	case "true":
		return nil
	}
	ip := ClientIP(r)
	isProxy, kind, err := IsProxy(ip)
	if err != nil {
		log.Printf("[PROXYCHECK] lookup failed ip=%s: %v", ip, err)
		return nil
	}
	if isProxy {
		log.Printf("[PROXYCHECK] blocked ip=%s type=%s", ip, kind)
		if kind == "" {
			kind = "VPN"
		}
		return fmt.Errorf("VPN/proxy detected (%s)", kind)
	}
	return nil
}
