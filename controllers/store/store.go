package store

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type StorefrontPrice struct {
	CurrencyType       int                 `json:"CurrencyType"`
	Price              int                 `json:"Price"`
	StorefrontSaleData *StorefrontSaleData `json:"StorefrontSaleData,omitempty"`
}

type StorefrontSaleData struct {
	SalePercent   int    `json:"SalePercent"`
	SaleStartDate string `json:"SaleStartDate"`
	SaleEndDate   string `json:"SaleEndDate"`
}

type StorefrontDrop struct {
	GiftDropId                int    `json:"GiftDropId"`
	FriendlyName              string `json:"FriendlyName"`
	Tooltip                   string `json:"Tooltip"`
	ConsumableItemDesc        string `json:"ConsumableItemDesc"`
	AvatarItemDesc            string `json:"AvatarItemDesc"`
	AvatarItemType            *int   `json:"AvatarItemType"`
	EquipmentPrefabName       string `json:"EquipmentPrefabName"`
	EquipmentModificationGuid string `json:"EquipmentModificationGuid"`
	IsQuery                   bool   `json:"IsQuery"`
	QueryRedirectContext      *int   `json:"QueryRedirectContext"`
	QueryRedirectRarity       *int   `json:"QueryRedirectRarity"`
	Unique                    bool   `json:"Unique"`
	SubscribersOnly           bool   `json:"SubscribersOnly"`
	Level                     int    `json:"Level,omitempty"`
	Rarity                    int    `json:"Rarity"`
	CurrencyType              int    `json:"CurrencyType"`
	Currency                  int    `json:"Currency"`
	Context                   int    `json:"Context"`
	ItemSetId                 *int   `json:"ItemSetId"`
	ItemSetFriendlyName       string `json:"ItemSetFriendlyName"`
}

type StorefrontItem struct {
	GiftDrop          *StorefrontDrop   `json:"GiftDrop,omitempty"`
	GiftDrops         []StorefrontDrop  `json:"GiftDrops,omitempty"`
	PurchasableItemId int               `json:"PurchasableItemId"`
	Type              int               `json:"Type"`
	Prices            []StorefrontPrice `json:"Prices"`
	SubscriberPrices  []StorefrontPrice `json:"SubscriberPrices"`
	IsFeatured        bool              `json:"IsFeatured"`
	NewUntil          *string           `json:"NewUntil"`
}

type Storefront struct {
	StoreItems                []StorefrontItem `json:"StoreItems"`
	StorefrontType            int              `json:"StorefrontType"`
	NextUpdate                string           `json:"NextUpdate"`
	NewUntil                  *string          `json:"NewUntil"`
	SubscriberDiscountPercent int              `json:"SubscriberDiscountPercent"`
}

const storefrontSeedDir = "db/seeds/storefronts"

func StorefrontDataDir() string {
	if d := os.Getenv("STOREFRONT_DATA_DIR"); d != "" {
		return d
	}
	return "data/storefronts"
}

func applyStorefrontDefaults(sf *Storefront) {
	if sf.NextUpdate == "" {
		sf.NextUpdate = "2222-02-22T22:22:00Z"
	}
	if sf.StoreItems == nil {
		sf.StoreItems = []StorefrontItem{}
	}
}

func loadStorefront(sfType int) (Storefront, bool) {
	name := fmt.Sprintf("sf%d.json", sfType)
	for _, dir := range []string{StorefrontDataDir(), storefrontSeedDir} {
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			continue
		}
		var sf Storefront
		if err := json.Unmarshal(data, &sf); err != nil {
			log.Printf("[STOREFRONTS] parse %s/%s: %v", dir, name, err)
			continue
		}
		applyStorefrontDefaults(&sf)
		return sf, true
	}
	return Storefront{}, false
}

func LookupStorefrontItem(purchasableItemID, storefrontType int) (StorefrontItem, int, bool) {
	if storefrontType != 0 {
		if sf, ok := loadStorefront(storefrontType); ok {
			for _, it := range sf.StoreItems {
				if it.PurchasableItemId == purchasableItemID {
					return it, sf.StorefrontType, true
				}
			}
		}
	}
	for _, sf := range AllStorefronts() {
		for _, it := range sf.StoreItems {
			if it.PurchasableItemId == purchasableItemID {
				return it, sf.StorefrontType, true
			}
		}
	}
	return StorefrontItem{}, 0, false
}

func AllStorefronts() []Storefront {
	byType := map[int]Storefront{}
	for _, dir := range []string{storefrontSeedDir, StorefrontDataDir()} {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), ".json") {
				continue
			}
			data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
			if err != nil {
				log.Printf("[STOREFRONTS] read %s/%s: %v", dir, entry.Name(), err)
				continue
			}
			var sf Storefront
			if err := json.Unmarshal(data, &sf); err != nil {
				log.Printf("[STOREFRONTS] parse %s/%s: %v", dir, entry.Name(), err)
				continue
			}
			if sf.StorefrontType == 0 {
				log.Printf("[STOREFRONTS] %s/%s missing StorefrontType, skipping", dir, entry.Name())
				continue
			}
			applyStorefrontDefaults(&sf)
			byType[sf.StorefrontType] = sf
		}
	}
	out := make([]Storefront, 0, len(byType))
	for _, sf := range byType {
		out = append(out, sf)
	}
	return out
}

func GetStorefront(sfType int) (Storefront, bool) {
	return loadStorefront(sfType)
}

func StorefrontByType(w http.ResponseWriter, r *http.Request) {
	log.Printf("[STOREFRONTS] %s", r.URL.Path)
	w.Header().Set("Content-Type", "application/json")

	parts := strings.Split(strings.TrimRight(r.URL.Path, "/"), "/")
	sfType, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if isStorefrontForcedEmpty(sfType) {
		log.Printf("[STOREFRONTS] type %d forced empty via STOREFRONT_EMPTY", sfType)
		w.Header().Set("X-Empty-Storefront", "true")
		writeEmptyStorefront(w, sfType)
		return
	}

	sf, ok := loadStorefront(sfType)
	if !ok {
		w.Header().Set("X-Empty-Storefront", "true")
		writeEmptyStorefront(w, sfType)
		return
	}
	json.NewEncoder(w).Encode(sf)
}

func writeEmptyStorefront(w http.ResponseWriter, sfType int) {
	json.NewEncoder(w).Encode(emptyStorefrontResponse{
		StorefrontType:            sfType,
		NextUpdate:                "2030-01-01T00:00:00Z",
		SubscriberDiscountPercent: 0,
		StoreItems:                []any{},
	})
}

func isStorefrontForcedEmpty(sfType int) bool {
	raw := os.Getenv("STOREFRONT_EMPTY")
	if raw == "" {
		return false
	}
	for part := range strings.SplitSeq(raw, ",") {
		if n, err := strconv.Atoi(strings.TrimSpace(part)); err == nil && n == sfType {
			return true
		}
	}
	return false
}

type emptyStorefrontResponse struct {
	StorefrontType            int    `json:"StorefrontType"`
	NextUpdate                string `json:"NextUpdate"`
	SubscriberDiscountPercent int    `json:"SubscriberDiscountPercent"`
	StoreItems                []any  `json:"StoreItems"`
}

func GameRewards(w http.ResponseWriter, r *http.Request) {
	log.Printf("[GAMEREWARDS] pending")
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("[]"))
}

func CampusCard(w http.ResponseWriter, r *http.Request) {
	log.Printf("[CAMPUSCARD] %s", r.URL.Path)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"platformAccountSubscribedPlayerId": nil,
		"subscription": map[string]any{
			"createdAt":          "0001-01-01T00:00:00",
			"expirationDate":     "2050-01-01T00:00:00Z",
			"isAutoRenewing":     true,
			"level":              0,
			"modifiedAt":         "0001-01-01T00:00:00",
			"period":             0,
			"platformId":         "",
			"platformPurchaseId": "",
			"platformType":       -1,
			"recNetPlayerId":     0,
			"subscriptionId":     0,
		},
	})
}
