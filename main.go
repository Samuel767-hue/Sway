package main

import (
	"fmt"
	"net/http"
	"os"
	"github.com/gorilla/websocket"
)

// Configuración del WebSocket
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer ws.Close()

	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			break
		}
		// Reenviamos el mensaje recibido
		err = ws.WriteMessage(1, msg)
		if err != nil {
			break
		}
	}
}

func main() {
	// 1. Servir el archivo index.html directamente cuando alguien entra a la raíz
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	// 2. Ruta del WebSocket
	http.HandleFunc("/ws", handleConnections)

	// 3. Obtener el puerto de Render o usar 8080 por defecto
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Println("Sway activo en el puerto:", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		fmt.Println("Error al iniciar el servidor:", err)
	}
}
