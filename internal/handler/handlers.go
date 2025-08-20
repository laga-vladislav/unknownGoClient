package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"reflect"

	"github.com/xtls/xray-core/app/proxyman/command"
	"github.com/xtls/xray-core/common/protocol"
	"github.com/xtls/xray-core/common/serial"
	"github.com/xtls/xray-core/proxy/vless"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type UserInfo struct {
	InTag  string `json:"in_tag"`
	Level  uint32 `json:"level"`
	Email  string `json:"email"`
	Uuid   string `json:"uuid"`
	Flow   string `json:"flow"`
}

func GetConfigHandler(w http.ResponseWriter, r *http.Request, configPath string) {
	log.Printf("GET /config requested, path=%s, client=%s", configPath, r.RemoteAddr)

	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Printf("Failed to read config: %v (path=%s)", err, configPath)
		http.Error(w, "Failed to read config", http.StatusInternalServerError)
		return
	}

	log.Printf("Config loaded successfully (size=%d bytes)", len(data))
	if len(data) > 200 {
		log.Printf("Preview (first 200 chars): %s...", string(data[:200]))
	} else {
		log.Printf("Preview: %s", string(data))
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(data); err != nil {
		log.Printf("Failed to write response: %v", err)
	} else {
		log.Printf("Response sent to client successfully")
	}
}

func PostConfigHandler(w http.ResponseWriter, r *http.Request, configPath string) {
	log.Printf("POST /config requested, path=%s, client=%s", configPath, r.RemoteAddr)

	if r.Header.Get("Content-Type") != "application/json" {
		log.Printf("Invalid Content-Type: %s (expected application/json)", r.Header.Get("Content-Type"))
		http.Error(w, "Content-Type must be application/json", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Failed to read request body: %v", err)
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	log.Printf("Request body received (size=%d bytes)", len(body))
	if len(body) > 200 {
		log.Printf("Preview (first 200 chars): %s...", string(body[:200]))
	} else {
		log.Printf("Preview: %s", string(body))
	}

	var tmp map[string]interface{}
	err = json.Unmarshal(body, &tmp)
	if err != nil {
		log.Printf("Invalid JSON format: %v", err)
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	log.Printf("Parsed JSON keys: %+v", reflect.ValueOf(tmp).MapKeys())

	if _, ok := tmp["inbounds"]; !ok {
		log.Print("Validation failed: missing 'inbounds'")
		http.Error(w, "Missing 'inbounds' in config", http.StatusBadRequest)
		return
	}
	if _, ok := tmp["outbounds"]; !ok {
		log.Print("Validation failed: missing 'outbounds'")
		http.Error(w, "Missing 'outbounds' in config", http.StatusBadRequest)
		return
	}

	err = os.WriteFile(configPath, body, 0644)
	if err != nil {
		log.Printf("Failed to write config to %s: %v", configPath, err)
		http.Error(w, "Failed to write config", http.StatusInternalServerError)
		return
	}

	log.Printf("Config written successfully to %s", configPath)
	w.WriteHeader(http.StatusAccepted)
	log.Print("Response status 202 Accepted sent to client")
}

func GetGrpcClient(xrayApiPort string) (command.HandlerServiceClient, *grpc.ClientConn, error) {
	log.Printf("Connecting to gRPC server at xray:%s", xrayApiPort)
	conn, err := grpc.NewClient(fmt.Sprintf("xray:%s", xrayApiPort), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("Failed to connect to gRPC server: %v", err)
		return nil, nil, err
	}
	client := command.NewHandlerServiceClient(conn)
	log.Printf("gRPC client connected successfully")
	return client, conn, nil
}

func AddVlessUser(client command.HandlerServiceClient, user *UserInfo) error {
	log.Printf("Adding VLESS user: in_tag=%s, email=%s, uuid=%s, flow=%s, level=%d", 
		user.InTag, user.Email, user.Uuid, user.Flow, user.Level)
	_, err := client.AlterInbound(context.Background(), &command.AlterInboundRequest{
		Tag: user.InTag,
		Operation: serial.ToTypedMessage(&command.AddUserOperation{
			User: &protocol.User{
				Level: user.Level,
				Email: user.Email,
				Account: serial.ToTypedMessage(&vless.Account{
					Id:   user.Uuid,
					Flow: user.Flow,
				}),
			},
		}),
	})
	if err != nil {
		log.Printf("Failed to add VLESS user: %v", err)
		return err
	}
	log.Printf("VLESS user added successfully")
	return nil
}

func RemoveVlessUser(client command.HandlerServiceClient, user *UserInfo) error {
	log.Printf("Removing VLESS user: in_tag=%s, email=%s", user.InTag, user.Email)
	_, err := client.AlterInbound(context.Background(), &command.AlterInboundRequest{
		Tag: user.InTag,
		Operation: serial.ToTypedMessage(&command.RemoveUserOperation{
			Email: user.Email,
		}),
	})
	if err != nil {
		log.Printf("Failed to remove VLESS user: %v", err)
		return err
	}
	log.Printf("VLESS user removed successfully")
	return nil
}
