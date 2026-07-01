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
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	clients = make(map[string]*websocket.Conn)
	mutex   = sync.Mutex{}
	history = make([]string, 0)
)

func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer ws.Close()

	ws.WriteMessage(1, []byte("SISTEMA: Conectado a los servidores de Sway"))

	_, p, _ := ws.ReadMessage()
	username := strings.TrimSpace(string(p))

	mutex.Lock()
	clients[username] = ws
	fmt.Println("LOG: Sesión iniciada por el usuario:", username)

	// Persistencia garantizada enviando el historial al reconectar
	for _, oldMessage := range history {
		ws.WriteMessage(1, []byte(oldMessage))
	}
	mutex.Unlock()

	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			mutex.Lock()
			delete(clients, username)
			fmt.Println("LOG: Conexión cerrada por el usuario:", username)
			mutex.Unlock()
			break
		}

		message := string(msg)
		fmt.Printf("LOG MESSAGE [%s]: %s\n", username, message)

		if strings.HasPrefix(message, "@") {
			parts := strings.SplitN(message, " ", 2)
			target := strings.TrimPrefix(parts[0], "@")

			mutex.Lock()
			if conn, ok := clients[target]; ok {
				if len(parts) > 1 {
					privateMessage := "(Privado de " + username + "): " + parts[1]
					history = append(history, privateMessage)
					
					conn.WriteMessage(1, []byte(privateMessage))
					ws.WriteMessage(1, []byte(privateMessage))
				} else {
					ws.WriteMessage(1, []byte("SISTEMA: Formato inválido. Escribe @usuario mensaje"))
				}
			} else {
				ws.WriteMessage(1, []byte("SISTEMA: El usuario '"+target+"' no se encuentra conectado"))
			}
			mutex.Unlock()

		} else {
			fullMessage := username + ": " + message

			mutex.Lock()
			history = append(history, fullMessage)

			for _, client := range clients {
				client.WriteMessage(1, []byte(fullMessage))
			}
			mutex.Unlock()
		}
	}
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})
	http.HandleFunc("/ws", handleConnections)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Println("Servidor Sway corriendo de forma segura en puerto:", port)
	http.ListenAndServe(":"+port, nil)
}
