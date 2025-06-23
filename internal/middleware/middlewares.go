package middleware

import (
	"net/http"
	"os"
	"strings"
	"log"
)


var allowedIP string
var apiToken string


func IpWhitelistMiddleware(next http.HandlerFunc) http.HandlerFunc {
	allowedIP = os.Getenv("ALLOWED_IP")
	if allowedIP == "" {
		log.Fatal("ALLOWED_IP is not set")
	}
	return func(w http.ResponseWriter, r *http.Request) {
		ip := strings.Split(r.RemoteAddr, ":")[0]
		log.Printf("Request from IP: %s", ip)
		if ip != allowedIP {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		next(w, r)
	}
}


func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	apiToken = os.Getenv("API_TOKEN")
	if apiToken == "" {
    	log.Fatal("API_TOKEN is not set")
	}
	return func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") || strings.TrimPrefix(auth, "Bearer ") != apiToken {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}
