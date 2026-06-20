package controllers

import (
	"net/http"
	"os"
)

const charadesIcebreakersPath = "data/jsons/icebreakers.json"
const charadesWordsPath = "data/jsons/charades.json"

func CharadesIcebreakers(w http.ResponseWriter, r *http.Request) {
	if _, ok := AccountIDFromRequest(r); !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if _, err := os.Stat(charadesIcebreakersPath); err != nil {
		w.Write([]byte("[]"))
		return
	}

	http.ServeFile(w, r, charadesIcebreakersPath)
}

func CharadesWords(w http.ResponseWriter, r *http.Request) {
	if _, ok := AccountIDFromRequest(r); !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if _, err := os.Stat(charadesWordsPath); err != nil {
		w.Write([]byte("[]"))
		return
	}

	http.ServeFile(w, r, charadesWordsPath)
}
