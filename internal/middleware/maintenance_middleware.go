package middleware

import (
	"encoding/json"
	"net/http"
)

func MaintenanceMiddleware(isMaintenance bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if isMaintenance {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusServiceUnavailable) 
				json.NewEncoder(w).Encode(map[string]string{
					"message": "Hệ thống đang bảo trì, vui lòng quay lại sau.",
				})
				
				return 
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK) // 200
			json.NewEncoder(w).Encode(map[string]string{
				"data": "Chào mừng đến với hệ thống đặt vé!",
			})
			return
		})
	}
}
