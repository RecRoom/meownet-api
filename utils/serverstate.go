package utils

import (
	"math"
	"os"
	"strings"
	"sync"
	"time"
)

func BaseHost() string {
	if h := os.Getenv("HOST"); h != "" {
		return strings.TrimRight(h, "/")
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}
	return "http://localhost" + port
}

func ShareBaseUrl() string {
	if v := os.Getenv("SHARE_BASE_URL"); v != "" {
		return v
	}
	return BaseHost() + "/{0}"
}

var (
	maintenanceMu     sync.RWMutex
	maintenanceTarget time.Time
)

func SetMaintenance(startsInMinutes int) {
	maintenanceMu.Lock()
	defer maintenanceMu.Unlock()
	if startsInMinutes <= 0 {
		maintenanceTarget = time.Time{}
		return
	}
	maintenanceTarget = time.Now().Add(time.Duration(startsInMinutes) * time.Minute)
}

func MaintenanceStartsInMinutes() int {
	maintenanceMu.RLock()
	defer maintenanceMu.RUnlock()
	if maintenanceTarget.IsZero() {
		return 0
	}
	remaining := time.Until(maintenanceTarget)
	if remaining <= 0 {
		return 0
	}
	return int(math.Ceil(remaining.Minutes()))
}
