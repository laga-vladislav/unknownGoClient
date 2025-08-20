package main

import (
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
    http.HandleFunc("/xray-api/user", middleware.AuthMiddleware(handler.XrayApiUserHandler))

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

func xrayApiHandler() {
}
