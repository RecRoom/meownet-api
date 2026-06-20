package routes

import (
	"net/http"

	"meow.net/controllers/admin"
	"meow.net/controllers/importer"
)

// all are gated via admin token
func RegisterAdminRoutes() {
	// Room importer
	http.HandleFunc("POST /import/upload", importer.Upload)

	// Storefronts
	http.HandleFunc("GET /admin/storefronts", admin.ListStorefronts)
	http.HandleFunc("POST /admin/storefronts/upload", admin.UploadStorefront)
	http.HandleFunc("GET /admin/storefronts/{type}", admin.GetStorefront)

	// Bans
	http.HandleFunc("GET /admin/bans", admin.ListBans)
	http.HandleFunc("POST /admin/bans", admin.CreateBan)
	http.HandleFunc("DELETE /admin/bans/{account_id}", admin.DeleteBan)
	http.HandleFunc("POST /anticheat/ban", admin.AnticheatBan)
	http.HandleFunc("POST /admin/forceclose", admin.ForceClose)
	http.HandleFunc("POST /admin/forcejoin", admin.ForceJoin)
	// Accounts
	http.HandleFunc("POST /admin/accounts", admin.ForceCreateAccount)
	http.HandleFunc("GET /admin/accounts/{id}", admin.GetAccountDetail)
	http.HandleFunc("PATCH /admin/accounts/{id}", admin.UpdateAccount)
	http.HandleFunc("POST /admin/accounts/{id}/balance/adjust", admin.AdjustBalance)
	http.HandleFunc("DELETE /admin/accounts/{id}/items/{desc}", admin.RevokeAvatarItem)
	http.HandleFunc("POST /admin/accounts/{id}/gift", admin.Gift)
	http.HandleFunc("POST /admin/gift/bulk", admin.GiftBulk)
	http.HandleFunc("POST /admin/accounts/{id}/progression", admin.SetProgression)
	http.HandleFunc("POST /admin/accounts/{id}/kick", admin.KickPlayer)
	http.HandleFunc("POST /admin/accounts/{id}/presence/refresh", admin.RefreshPresence)

	// Messaging
	http.HandleFunc("POST /admin/messages/coach", admin.SendCoachMessage)
	http.HandleFunc("POST /admin/messages/coach/all", admin.SendCoachMessageAll)

	// Server maintenance broadcast
	http.HandleFunc("POST /admin/maintenance", admin.BroadcastServerMaintenance)

	// Reports
	http.HandleFunc("POST /admin/reports/act", admin.ActOnReport)

	// Stats
	http.HandleFunc("GET /admin/stats/players", admin.GetPlayerCount)

	// Instances
	http.HandleFunc("GET /admin/instances", admin.ListInstances)
	http.HandleFunc("POST /admin/instances/{id}/kill", admin.KillInstance)
}
