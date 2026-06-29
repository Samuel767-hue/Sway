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
)

func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer ws.Close()

	// Registro de usuario
	ws.WriteMessage(1, []byte("SISTEMA: Escribe tu nombre para entrar"))
	_, p, _ := ws.ReadMessage()
	username := strings.TrimSpace(string(p))

	mutex.Lock()
	clients[username] = ws
	fmt.Println("LOG: Usuario conectado:", username)
	mutex.Unlock()

	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			mutex.Lock()
			delete(clients, username)
			fmt.Println("LOG: Usuario desconectado:", username)
			mutex.Unlock()
			break
		}

		message := string(msg)
		fmt.Println("LOG: Mensaje de", username, ":", message)

		if strings.HasPrefix(message, "@") {
			parts := strings.SplitN(message, " ", 2)
			target := strings.TrimPrefix(parts[0], "@")

			mutex.Lock()
			if conn, ok := clients[target]; ok {
				conn.WriteMessage(1, []byte("(Privado de "+username+"): "+parts[1]))
				fmt.Println("LOG: Mensaje privado enviado a", target)
			} else {
				ws.WriteMessage(1, []byte("SISTEMA: Usuario "+target+" no encontrado"))
			}
			mutex.Unlock()
		} else {
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
	if port == "" {
		port = "8080"
	}
	fmt.Println("Sway activo en puerto:", port)
	http.ListenAndServe(":"+port, nil)
}
