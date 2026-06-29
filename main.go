package main

import (
	"fmt"
	"net/http"
	"os"
	"sync"
	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	clients   = make(map[*websocket.Conn]bool)
	broadcast = make(chan []byte)
	mutex     = sync.Mutex{}
)

func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, _ := upgrader.Upgrade(w, r, nil)
	defer ws.Close()

	mutex.Lock()
	clients[ws] = true
	mutex.Unlock()

	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			mutex.Lock()
			delete(clients, ws)
			mutex.Unlock()
			break
		}
		broadcast <- msg
	}
}

func handleMessages() {
	for {
		msg := <-broadcast
		mutex.Lock()
		for client := range clients {
			err := client.WriteMessage(1, msg)
			if err != nil {
				client.Close()
				delete(clients, client)
			}
		}
		mutex.Unlock()
	}
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, "index.html") })
	http.HandleFunc("/ws", handleConnections)

	go handleMessages()

	port := os.Getenv("PORT")
	if port == "" { port = "8080" }
	
	fmt.Println("Sway activo en puerto:", port)
	http.ListenAndServe(":"+port, nil)
}
