package main

import (
    "fmt"
    "net/http"
    "github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool { return true }, 
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
    ws, _ := upgrader.Upgrade(w, r, nil)
    defer ws.Close()

    for {
        // Leer mensaje del cliente
        _, msg, err := ws.ReadMessage()
        if err != nil {
            break
        }
        fmt.Printf("Mensaje recibido: %s\n", msg)
        
        // Devolver mensaje (Eco)
        err = ws.WriteMessage(1, msg)
        if err != nil {
            break
        }
    }
}

func main() {
    http.HandleFunc("/ws", handleConnections)
    fmt.Println("Servidor iniciado en :8080")
    http.ListenAndServe(":8080", nil)
}