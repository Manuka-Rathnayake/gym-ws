package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

var clients = make(map[string]*websocket.Conn)
var mutex = &sync.Mutex{}

type RegisterMessage struct {
	Type  string `json:"type"`
	GymID string `json:"gymID"`
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("WebSocket Upgrade Error:", err)
		return
	}

	defer conn.Close()

	// Expect JSON payload from ESP (gymID registration)
	_, msg, err := conn.ReadMessage()
	if err != nil {
		fmt.Println("Error reading gym ID:", err)
		return
	}

	var registerMsg RegisterMessage
	err = json.Unmarshal(msg, &registerMsg)
	if err != nil || registerMsg.Type != "REGISTER" || registerMsg.GymID == "" {
		fmt.Println("Invalid registration message:", string(msg))
		return
	}

	gymID := registerMsg.GymID
	fmt.Printf("✅ ESP connected for Gym ID: %s\n", gymID)

	// Store the ESP connection
	mutex.Lock()
	clients[gymID] = conn
	mutex.Unlock()

	// Keep WebSocket connection alive
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			fmt.Printf("⚠️ ESP for Gym ID %s disconnected: %v\n", gymID, err)
			mutex.Lock()
			delete(clients, gymID)
			mutex.Unlock()
			break
		}
	}
}

func unlockDoor(w http.ResponseWriter, r *http.Request) {
	gymID := r.URL.Query().Get("gymID") // Expecting gym ID as a query parameter
	if gymID == "" {
		http.Error(w, "Missing gymID parameter", http.StatusBadRequest)
		return
	}

	mutex.Lock()
	conn, exists := clients[gymID]
	mutex.Unlock()

	if !exists {
		http.Error(w, "ESP not connected", http.StatusNotFound)
		fmt.Printf("❌ ESP not connected for Gym ID: %s\n", gymID)
		return
	}

	err := conn.WriteMessage(websocket.TextMessage, []byte("UNLOCK"))
	if err != nil {
		fmt.Printf("❌ Error sending unlock signal to Gym ID %s: %v\n", gymID, err)
		http.Error(w, "Failed to send unlock command", http.StatusInternalServerError)
		return
	}

	fmt.Printf("✅ Unlock signal sent to Gym ID: %s\n", gymID)
	fmt.Fprintln(w, "✅ Unlock signal sent")
}

func main() {
	http.HandleFunc("/ws", handleConnections)
	http.HandleFunc("/unlock", unlockDoor)

	fmt.Println("✅ WebSocket Server started on port 8080")
	http.ListenAndServe(":8080", nil)
}
