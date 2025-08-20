package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"unknowngoclient/internal/handler"
	"unknowngoclient/internal/middleware"

	"github.com/joho/godotenv"
)

var configPath string
var xrayApiPort string

func init() {
    err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file")
    }

    configPath = os.Getenv("XRAY_CONFIG_DIR")
    xrayApiPort = os.Getenv("XRAY_API_PORT")
    if configPath == "" || xrayApiPort == "" {
        log.Fatal("XRAY_CONFIG_DIR or XRAY_API_PORT are not set")
    } else {
        configPath += "/config.json"
    }
    log.Println(configPath)
    log.Println(xrayApiPort)
}

func main() {
    http.HandleFunc("/config", middleware.AuthMiddleware(configHandler))
    http.HandleFunc("/xray-api/user", middleware.AuthMiddleware(XrayApiUserHandler))

    port := os.Getenv("INTERNAL_SERVER_PORT")
    log.Printf("Server listening on :%s", port)
    log.Fatal(http.ListenAndServe(":"+port, nil))
}

func configHandler(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodGet:
        handler.GetConfigHandler(w, r, configPath)
    case http.MethodPost:
        handler.PostConfigHandler(w, r, configPath)
    default:
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    }
}

func XrayApiUserHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received %s request to /xray-api/user from %s", r.Method, r.RemoteAddr)
	switch r.Method {
	case http.MethodGet:
		log.Printf("Handling GET /xray-api/user")
		handler.GetConfigHandler(w, r, os.Getenv("XRAY_CONFIG_DIR")+"/config.json")
	case http.MethodPost:
		log.Printf("Handling POST /xray-api/user")
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("Failed to read request body: %v", err)
			http.Error(w, "Failed to read body", http.StatusInternalServerError)
			return
		}
		log.Printf("Request body received (size=%d bytes): %s", len(body), string(body))
		var user handler.UserInfo
		if err := json.Unmarshal(body, &user); err != nil {
			log.Printf("Invalid JSON: %v", err)
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		log.Printf("Parsed user data: %+v", user)
		if user.InTag == "" || user.Email == "" || user.Uuid == "" {
			log.Printf("Validation failed: missing required fields (in_tag=%s, email=%s, uuid=%s)", 
				user.InTag, user.Email, user.Uuid)
			http.Error(w, "Missing fields", http.StatusBadRequest)
			return
		}
		client, conn, err := handler.GetGrpcClient(xrayApiPort)
		if err != nil {
			log.Printf("gRPC connect failed: %v", err)
			http.Error(w, "gRPC connect failed", http.StatusInternalServerError)
			return
		}
		defer conn.Close()
		if err := handler.AddVlessUser(client, &user); err != nil {
			log.Printf("AddVlessUser failed: %v", err)
			http.Error(w, fmt.Sprintf("Add user failed: %v", err), http.StatusInternalServerError)
			return
		}
		log.Printf("User added successfully, responding with 202 Accepted")
		w.WriteHeader(http.StatusAccepted)
	default:
		log.Printf("Invalid method %s, responding with 405 Method Not Allowed", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}