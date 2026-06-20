package routes

import (
	"net/http"

	"meow.net/controllers"
)

func RegisterPatcherRoutes() {
	http.HandleFunc("GET /patcher/version", controllers.PatcherVersion)
	http.HandleFunc("/patcher/validate", controllers.PatcherValidate)
}
