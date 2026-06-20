package controllers

import (
	"encoding/json"
	"net/http"
	"os"
	"sync"

	"meow.net/utils"
)

func GameConfigs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	http.ServeFile(w, r, "data/jsons/gameconfigs.json")
}

func LoadingScreenTips(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	http.ServeFile(w, r, "data/jsons/loadingscreentipdata.json")
}

const configV2Path = "data/jsons/configv2.json"

var (
	configV2Once sync.Once
	configV2Base map[string]json.RawMessage
	configV2Err  error
)

func loadConfigV2Base() (map[string]json.RawMessage, error) {
	configV2Once.Do(func() {
		data, err := os.ReadFile(configV2Path)
		if err != nil {
			configV2Err = err
			return
		}
		configV2Err = json.Unmarshal(data, &configV2Base)
	})
	return configV2Base, configV2Err
}

func ConfigV2(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	base, err := loadConfigV2Base()
	if err != nil {
		http.Error(w, "config unavailable", http.StatusInternalServerError)
		return
	}

	out := make(map[string]json.RawMessage, len(base)+2)
	for k, v := range base {
		out[k] = v
	}

	if maint, err := json.Marshal(map[string]int{
		"StartsInMinutes": utils.MaintenanceStartsInMinutes(),
	}); err == nil {
		out["ServerMaintenance"] = maint
	}

	if share, err := json.Marshal(utils.ShareBaseUrl()); err == nil {
		out["ShareBaseUrl"] = share
	}

	json.NewEncoder(w).Encode(out)
}
