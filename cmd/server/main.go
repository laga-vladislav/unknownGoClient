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

func init() {
    err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file")
    }

    configPath = os.Getenv("XRAY_CONFIG_PATH")
    if configPath == "" {
        log.Fatal("XRAY_CONFIG_PATH is not set")
    }
    log.Println(configPath)
}

func main() {
    http.HandleFunc("/config",
        middleware.IpWhitelistMiddleware(
            middleware.AuthMiddleware(
                configHandler,
            ),
        ),
    )

    port := os.Getenv("PORT")
    if port == "" {
        port = "7342"
    }
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
