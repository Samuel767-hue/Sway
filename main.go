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

	// Mensaje inicial del sistema
	ws.WriteMessage(1, []byte("SISTEMA: Escribe tu nombre para entrar"))

	_, p, _ := ws.ReadMessage()
	username := strings.TrimSpace(string(p))

	mutex.Lock()
	clients[username] = ws
	fmt.Println("LOG: Usuario conectado:", username)

	// Enviar historial completo al entrar (comunidad + privados guardados)
	for _, oldMessage := range history {
		ws.WriteMessage(1, []byte(oldMessage))
	}
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
		fmt.Println("LOG:", username, ":", message)

		// MENSAJES PRIVADOS
		if strings.HasPrefix(message, "@") {
			parts := strings.SplitN(message, " ", 2)
			target := strings.TrimPrefix(parts[0], "@")

			mutex.Lock()
			if conn, ok := clients[target]; ok {
				if len(parts) > 1 {
					privateMessage := "(Privado de " + username + "): " + parts[1]

					// Mantenemos tu guardado en el historial para que no se pierdan al recargar
					history = append(history, privateMessage)

					// Enviar al destinatario y a ti mismo para que aparezca en pantalla
					conn.WriteMessage(1, []byte(privateMessage))
					ws.WriteMessage(1, []byte(privateMessage))

					fmt.Println("LOG: Privado enviado a", target)
				} else {
					ws.WriteMessage(1, []byte("SISTEMA: Error formato @usuario mensaje"))
				}
			} else {
				ws.WriteMessage(1, []byte("SISTEMA: Usuario "+target+" no encontrado"))
			}
			mutex.Unlock()

		} else {
			// MENSAJE COMUNIDAD
			fullMessage := username + ": " + message

			mutex.Lock()
			history = append(history, fullMessage)

			// Enviar a todos
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

	fmt.Println("Sway activo en puerto:", port)
	http.ListenAndServe(":"+port, nil)
}
