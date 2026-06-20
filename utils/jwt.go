package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func jwtIssuer() string {
	if h := os.Getenv("HOST"); h != "" {
		return h
	}
	return "http://localhost:8080"
}

var rsaKey *rsa.PrivateKey

var rsaKid string

const rsaKeyFile = "jwt_key.pem"

func init() {
	if data, err := os.ReadFile(rsaKeyFile); err == nil {
		block, _ := pem.Decode(data)
		if block != nil {
			if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
				rsaKey = key
				rsaKid = deriveKid(&key.PublicKey)
				return
			}
		}
		log.Printf("[JWT] warning: could not parse %s, generating new key", rsaKeyFile)
	}

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}
	rsaKey = key
	rsaKid = deriveKid(&key.PublicKey)

	pemBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})
	if err := os.WriteFile(rsaKeyFile, pemBytes, 0600); err != nil {
		log.Printf("[JWT] warning: could not persist key to %s: %v", rsaKeyFile, err)
	}
}

func deriveKid(pub *rsa.PublicKey) string {
	der, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		der = pub.N.Bytes()
	}
	sum := sha256.Sum256(der)
	return strings.ToUpper(hex.EncodeToString(sum[:]))
}

func MakeJWT(accountId string, platformId string, platform string, isJunior bool, isDeveloper bool, isModerator bool) string {
	now := time.Now().UTC()
	exp := now.Add(12 * time.Hour)

	if platform == "" {
		platform = "0"
	}

	jti := strings.ToUpper(strings.ReplaceAll(uuid.New().String(), "-", ""))
	if len(jti) > 32 {
		jti = jti[:32]
	}

	roles := []string{"screenshare", "gameClient"}
	if isJunior {
		roles = []string{"junior", "gameClient"}
	}
	if isDeveloper {
		roles = append(roles, "developer", "moderator", "betaroomcurrencycreator")
	} else if isModerator {
		roles = append(roles, "moderator")
	}

	claims := jwt.MapClaims{
		"iss":       jwtIssuer(),
		"nbf":       now.Unix(),
		"iat":       now.Unix(),
		"exp":       exp.Unix(),
		"aud":       jwtIssuer(),
		"client_id": "rec",
		"sub":       accountId,
		"rn.plat":   platform,
		"rn.platid": platformId,
		"role":      roles,
		"jti":       jti,
		"scope": []string{
			"openid", "rn.api", "rn.commerce", "rn.notify",
			"rn.match", "rn.chat", "rn.accounts", "rn.auth",
			"rn.link", "rn.lists", "rn.clubs", "rn.rooms",
			"rn.data", "offline_access",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = rsaKid

	tokenString, err := token.SignedString(rsaKey)
	if err != nil {
		fmt.Printf("Error signing JWT: %v\n", err)
		return ""
	}

	return tokenString
}

type cachedSub struct {
	sub string
	iat time.Time
	exp time.Time
}

var (
	jwtSubCache   = map[string]cachedSub{}
	jwtSubCacheMu sync.RWMutex
)

const jwtCacheMaxEntries = 100000

var (
	revokedBefore   = map[string]time.Time{}
	revokedBeforeMu sync.RWMutex
)

func RevokeAccessTokens(accountId string) {
	now := time.Now()

	revokedBeforeMu.Lock()
	revokedBefore[accountId] = now
	for sub, cutoff := range revokedBefore {
		if now.Sub(cutoff) > 12*time.Hour {
			delete(revokedBefore, sub)
		}
	}
	revokedBeforeMu.Unlock()
}

func accessRevoked(sub string, iat time.Time) bool {
	revokedBeforeMu.RLock()
	cutoff, ok := revokedBefore[sub]
	revokedBeforeMu.RUnlock()
	return ok && iat.Before(cutoff)
}

func ParseSubFromJWT(tokenStr string) (string, error) {
	now := time.Now()

	jwtSubCacheMu.RLock()
	c, ok := jwtSubCache[tokenStr]
	jwtSubCacheMu.RUnlock()
	if ok && now.Before(c.exp) {
		if accessRevoked(c.sub, c.iat) {
			return "", fmt.Errorf("invalid token")
		}
		return c.sub, nil
	}

	token, err := jwt.ParseWithClaims(tokenStr, jwt.MapClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return &rsaKey.PublicKey, nil
	})
	if err != nil || !token.Valid {
		return "", fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("invalid claims")
	}

	sub, ok := claims["sub"].(string)
	if !ok {
		return "", fmt.Errorf("no sub claim")
	}

	iat := now
	if i, err := claims.GetIssuedAt(); err == nil && i != nil {
		iat = i.Time
	}
	if accessRevoked(sub, iat) {
		return "", fmt.Errorf("invalid token")
	}

	exp := now.Add(time.Hour)
	if e, err := claims.GetExpirationTime(); err == nil && e != nil {
		exp = e.Time
	}

	jwtSubCacheMu.Lock()
	if len(jwtSubCache) >= jwtCacheMaxEntries {
		for k, v := range jwtSubCache {
			if now.After(v.exp) {
				delete(jwtSubCache, k)
			}
		}
	}
	jwtSubCache[tokenStr] = cachedSub{sub: sub, iat: iat, exp: exp}
	jwtSubCacheMu.Unlock()

	return sub, nil
}
