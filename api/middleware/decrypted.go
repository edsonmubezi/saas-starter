package middleware

import (
	"log"
	"net/http"

	secureid "github.com/edsonmubezi/myapp/pkg/encrypt"

	"github.com/gorilla/mux"
)

func DecryptMiddleware(paramKeys ...string) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			vars := mux.Vars(r)

			if len(vars) == 0 {
				// No path variables at all — skip decryption entirely
				next.ServeHTTP(w, r)
				return
			}

			for _, key := range paramKeys {
				if encryptedVal, exists := vars[key]; exists {
					decryptedVal, err := secureid.DecryptID(encryptedVal)
					if err != nil {
						// SECURITY: Do not log encrypted values as they may be sensitive
						log.Printf("[Decryption Error] Param: %s | Error: %v", key, err)
						http.Error(w, "Invalid or tampered ID", http.StatusBadRequest)
						return
					}
					vars[key] = decryptedVal
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
