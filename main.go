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
	clients  = make(map[string]*websocket.Conn) // Mapa: nombre -> conexión
	mutex    = sync.Mutex{}
)

func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, _ := upgrader.Upgrade(w, r, nil)
	defer ws.Close()

	// 1. Pedimos el nombre al conectar (lo enviamos como primer mensaje)
	ws.WriteMessage(1, []byte("SISTEMA: Escribe tu nombre para empezar"))
	_, p, _ := ws.ReadMessage()
	username := string(p)

	mutex.Lock()
	clients[username] = ws
	mutex.Unlock()

	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			mutex.Lock()
			delete(clients, username)
			mutex.Unlock()
			break
		}

		message := string(msg)
		// Si empieza por @, es privado
		if strings.HasPrefix(message, "@") {
			parts := strings.SplitN(message, " ", 2)
			target := strings.TrimPrefix(parts[0], "@")
			
			mutex.Lock()
			if conn, ok := clients[target]; ok {
				conn.WriteMessage(1, []byte("(Privado de "+username+"): "+parts[1]))
			}
			mutex.Unlock()
		} else {
			// Si no, es público
			mutex.Lock()
			for _, client := range clients {
				client.WriteMessage(1, []byte(username+": "+message))
			}
			mutex.Unlock()
		}
	}
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, "index.html") })
	http.HandleFunc("/ws", handleConnections)
	port := os.Getenv("PORT")
	if port == "" { port = "8080" }
	http.ListenAndServe(":"+port, nil)
}
