package rooms

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func Rooms(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"RoomName":         "RecRoom",
		"IsDormRoom":       true,
		"RoomInstanceId":   0,
		"PhotonRoomName":   nil,
		"PhotonRegion":     nil,
		"PhotonAppVersion": nil,
	})
}

func RoomCurrencies(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if strings.HasSuffix(r.URL.Path, "/betaEnabled") {
		fmt.Fprint(w, "true")
	} else {
		fmt.Fprint(w, "[]")
	}
}

func RoomKeys(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, "[]")
}
