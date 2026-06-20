package controllers

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os"
)

func PatcherVersion(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	http.ServeFile(w, r, "data/jsons/patcher_version.json")
}

func PatcherValidate(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Version string `json:"version"`
		Key     string `json:"key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("false"))
		return
	}

	versionFile, err := os.ReadFile("data/jsons/patcher_version.json")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("false"))
		return
	}
	var versionData struct {
		Version string `json:"Version"`
	}
	if err := json.Unmarshal(versionFile, &versionData); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("false"))
		return
	}

	seed, err := base64.StdEncoding.DecodeString(os.Getenv("PATCHER_PRIVATE_KEY"))
	if err != nil || len(seed) != ed25519.SeedSize {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("false"))
		return
	}
	privateKey := ed25519.NewKeyFromSeed(seed)
	expectedPubKey := hex.EncodeToString(privateKey.Public().(ed25519.PublicKey))

	if body.Version != versionData.Version || body.Key != expectedPubKey {
		w.Write([]byte("false"))
		return
	}

	w.Write([]byte("true"))
}
