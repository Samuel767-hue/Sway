package main

import (
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	clients  = make(map[string]*websocket.Conn)
	mutex    = sync.Mutex{}
)

func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, _ := upgrader.Upgrade(w, r, nil)
	defer ws.Close()

	_, p, _ := ws.ReadMessage()
	username := strings.TrimSpace(string(p))
	mutex.Lock()
	clients[username] = ws
	mutex.Unlock()

	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			break
		}

		text := string(msg)
		if strings.HasPrefix(text, "@") {
			// Lógica privado
			parts := strings.SplitN(text, " ", 2)
			target := strings.TrimPrefix(parts[0], "@")
			if conn, ok := clients[target]; ok {
				conn.WriteMessage(1, []byte(username+": "+parts[1]))
			}
		} else {
			// Lógica comunidad
			for _, client := range clients {
				client.WriteMessage(1, []byte(username+": "+text))
			}
		}
	}
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, "index.html") })
	http.HandleFunc("/ws", handleConnections)
	http.ListenAndServe(":8080", nil)
}
