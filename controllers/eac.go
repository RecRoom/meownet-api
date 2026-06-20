package controllers

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
)

func EACChallenge(w http.ResponseWriter, r *http.Request) {
	challengeToken := fmt.Sprintf("\"%s\"", uuid.New().String())

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintf(w, "%s", challengeToken)
}
