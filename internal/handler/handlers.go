package handler

import (
	"encoding/json"
	"io"
	"log"
	// "io/ioutil"
	"net/http"
	"os"
)

func GetConfigHandler(w http.ResponseWriter, r *http.Request, configPath string) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Printf("%s", string(data))
		http.Error(w, "Failed to read config", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func PostConfigHandler(w http.ResponseWriter, r *http.Request, configPath string) {
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Content-Type must be application/json", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}

	var tmp map[string]interface{}
	err = json.Unmarshal(body, &tmp)
	if err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	log.Println(tmp)

	if _, ok := tmp["inbounds"]; !ok {
		http.Error(w, "Missing 'inbounds' in config", http.StatusBadRequest)
		return
	}
	if _, ok := tmp["outbounds"]; !ok {
		http.Error(w, "Missing 'outbounds' in config", http.StatusBadRequest)
		return
	}

	err = os.WriteFile(configPath, body, 0644)
	if err != nil {
		http.Error(w, "Failed to write config", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}