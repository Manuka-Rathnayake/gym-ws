package main

import (
	"fmt"
	"net/http"
	"sync"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // Allow all connections
}

var clients = make(map[string]*websocket.Conn) // Store ESP devices by gym ID
var mutex = &sync.Mutex{}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("WebSocket Upgrade Error:", err)
		return
	}

	defer conn.Close()

	// Expect the first message to contain the gym ID
	_, msg, err := conn.ReadMessage()
	if err != nil {
		fmt.Println("Error reading gym ID:", err)
		return
	}

	gymID := string(msg)
	fmt.Println("ESP connected for Gym ID:", gymID)

	// Store the ESP connection
	mutex.Lock()
	clients[gymID] = conn
	mutex.Unlock()

	// Keep connection alive
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("ESP disconnected:", err)
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
		return
	}

	err := conn.WriteMessage(websocket.TextMessage, []byte("UNLOCK"))
	if err != nil {
		fmt.Println("Error sending unlock signal:", err)
		http.Error(w, "Failed to send unlock command", http.StatusInternalServerError)
		return
	}

	fmt.Fprintln(w, "Unlock signal sent to gym:", gymID)
}

func main() {
	http.HandleFunc("/ws", handleConnections)
	http.HandleFunc("/unlock", unlockDoor)

	fmt.Println("WebSocket Server started on port 8080")
	http.ListenAndServe(":8080", nil)
}
