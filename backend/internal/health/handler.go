package health

import (
	"net/http"

	"github.com/freel/backend/internal/utils"
)

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Freel backend is running",
	})
}
