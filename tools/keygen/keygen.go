package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

func main() {
	seed := make([]byte, ed25519.SeedSize)
	if _, err := rand.Read(seed); err != nil {
		panic(err)
	}

	privateKey := ed25519.NewKeyFromSeed(seed)
	pubKey := privateKey.Public().(ed25519.PublicKey)

	fmt.Println("--- FOR SERVER (.env) ---")
	fmt.Println("PATCHER_PRIVATE_KEY=" + base64.StdEncoding.EncodeToString(seed))
	fmt.Println()
	fmt.Println("--- FOR CLIENT (JSON Payload) ---")
	fmt.Println("\"key\": \"" + hex.EncodeToString(pubKey) + "\"")
}
