package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Client struct {
	conn *websocket.Conn
	name string
}

var (
	clients   = make(map[*websocket.Conn]*Client)
	clientsMu sync.Mutex
)

func main() {
	// Servir archivos estáticos (index.html)
	http.Handle("/", http.FileServer(http.Dir("./")))
	http.HandleFunc("/ws", handleConnections)

	fmt.Println("Servidor corriendo en el puerto :8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error al mejorar la conexión:", err)
		return
	}
	defer ws.Close()

	// 1. Primer mensaje: Registrar el nombre del usuario
	_, nameBuf, err := ws.ReadMessage()
	if err != nil {
		return
	}
	username := strings.TrimSpace(string(nameBuf))
	if username == "" {
		username = "Invitado"
	}

	clientsMu.Lock()
	clients[ws] = &Client{conn: ws, name: username}
	clientsMu.Unlock()

	log.Printf("LOG: Usuario conectado: %s\n", username)

	// 2. Bucle principal de mensajes
	for {
		_, msgBuf, err := ws.ReadMessage()
		if err != nil {
			clientsMu.Lock()
			log.Printf("LOG: Usuario desconectado: %s\n", clients[ws].name)
			delete(clients, ws)
			clientsMu.Unlock()
			break
		}

		msgText := strings.TrimSpace(string(msgBuf))
		if msgText == "" {
			continue
		}

		// Sistema de Mensajes Privados (@usuario mensaje)
		if strings.HasPrefix(msgText, "@") {
			partes := strings.SplitN(msgText, " ", 2)
			
			// --- CAMBIO DE SEGURIDAD (ANTIBOMBAS) ---
			// Si len(partes) es menor que 2, significa que enviaron solo el @nombre sin texto
			if len(partes) < 2 {
				continue // Ignora el mensaje incompleto de forma segura en vez de romper el servidor
			}

			targetUser := strings.TrimPrefix(partes[0], "@")
			mensajeReal := partes[1]

			enviado := false
			clientsMu.Lock()
			for _, client := range clients {
				if client.name == targetUser {
					// Enviamos el privado manteniendo la estructura limpia
					client.conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("(Privado de %s): %s", username, mensajeReal)))
					enviado = true
					break
				}
			}
			clientsMu.Unlock()

			if !enviado && !strings.HasPrefix(mensajeReal, "/SYS") {
				ws.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("SISTEMA: El usuario %s no está conectado.", targetUser)))
			}
			continue
		}

		// Mensaje Público a la Comunidad
		log.Printf("LOG: Mensaje de %s: %s\n", username, msgText)
		clientsMu.Lock()
		for _, client := range clients {
			client.conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("%s: %s", username, msgText)))
		}
		clientsMu.Unlock()
	}
}
