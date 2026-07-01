package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	clients  = make(map[string]*websocket.Conn)
	mutex    = sync.Mutex{}
	history  = make([]string, 0)
)

func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil { return }
	defer ws.Close()

	// 1. Identificación del usuario
	_, p, _ := ws.ReadMessage()
	username := strings.TrimSpace(string(p))
	
	mutex.Lock()
	clients[username] = ws
	// Enviar historial previo al conectar
	for _, msg := range history {
		ws.WriteMessage(1, []byte(msg))
	}
	mutex.Unlock()

	// 2. Bucle de mensajes
	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			mutex.Lock()
			delete(clients, username)
			mutex.Unlock()
			break
		}
		
		text := string(msg)
		mutex.Lock()
		
		if strings.HasPrefix(text, "@") {
			parts := strings.SplitN(text, " ", 2)
			target := strings.TrimPrefix(parts[0], "@")
			if conn, ok := clients[target]; ok && len(parts) > 1 {
				priv := "(Privado de " + username + "): " + parts[1]
				conn.WriteMessage(1, []byte(priv))
				ws.WriteMessage(1, []byte(priv))
			}
		} else {
			full := username + ": " + text
			history = append(history, full)
			for _, client := range clients {
				client.WriteMessage(1, []byte(full))
			}
		}
		mutex.Unlock()
	}
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, "index.html") })
	http.HandleFunc("/ws", handleConnections)
	port := os.Getenv("PORT")
	if port == "" { port = "8080" }
	http.ListenAndServe(":"+port, nil)
}
