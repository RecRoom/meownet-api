package admin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"meow.net/controllers/store"
)

func ListStorefronts(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	writeJSON(w, http.StatusOK, store.AllStorefronts())
}

func GetStorefront(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	sfType, err := strconv.Atoi(r.PathValue("type"))
	if err != nil {
		http.Error(w, "invalid storefront type", http.StatusBadRequest)
		return
	}
	sf, ok := store.GetStorefront(sfType)
	if !ok {
		http.Error(w, "storefront not found", http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, sf)
}

func UploadStorefront(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}

	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 8<<20))
	if err != nil {
		http.Error(w, "failed to read body: "+err.Error(), http.StatusBadRequest)
		return
	}

	var sf store.Storefront
	if err := json.Unmarshal(body, &sf); err != nil {
		http.Error(w, "invalid storefront JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	if sf.StorefrontType == 0 {
		http.Error(w, "StorefrontType is required and must be non-zero", http.StatusBadRequest)
		return
	}

	var pretty bytes.Buffer
	if err := json.Indent(&pretty, body, "", "    "); err != nil {
		http.Error(w, "failed to format JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	dir := store.StorefrontDataDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		http.Error(w, "failed to create data dir: "+err.Error(), http.StatusInternalServerError)
		return
	}
	path := filepath.Join(dir, fmt.Sprintf("sf%d.json", sf.StorefrontType))
	if err := os.WriteFile(path, pretty.Bytes(), 0644); err != nil {
		http.Error(w, "failed to write file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success":         true,
		"storefront_type": sf.StorefrontType,
		"items":           len(sf.StoreItems),
		"path":            path,
	})
}
