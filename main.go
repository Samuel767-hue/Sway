package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, _ := upgrader.Upgrade(w, r, nil)
	defer ws.Close()

	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			break
		}
		fmt.Printf("Sway recibió: %s\n", msg)
		// Reenviamos el mensaje a quien esté conectado
		err = ws.WriteMessage(1, msg)
		if err != nil {
			break
		}
	}
}

func main() {
	// 1. Decirle a Go que sirva los archivos de la carpeta actual
	fs := http.FileServer(http.Dir("./"))
	http.Handle("/", fs)

	// 2. Tu ruta de WebSocket sigue igual
	http.HandleFunc("/ws", handleConnections)

	fmt.Println("Sway está activo en el puerto :8080")

	// 3. Ajuste importante para Render (usa la variable de entorno PORT)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	http.ListenAndServe(":"+port, nil)
}
