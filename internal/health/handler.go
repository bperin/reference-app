package health

import (
	"net/http"

	"github.com/example/reference-app/internal/http/response"
)

func Handler(w http.ResponseWriter, _ *http.Request) {
	response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
