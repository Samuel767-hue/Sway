package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	clients = make(map[string]*websocket.Conn)

	mutex = sync.Mutex{}
)

func saveCommunityMessage(username string, message string) {

	_, err := db.Exec(
		"INSERT INTO messages(username,message,type) VALUES(?,?,?)",
		username,
		message,
		"community",
	)

	if err != nil {
		fmt.Println("Error SQLite:", err)
	}

}

func sendHistory(ws *websocket.Conn) {

	rows, err := db.Query(
		"SELECT username,message FROM messages ORDER BY id ASC",
	)

	if err != nil {
		fmt.Println("Error leyendo historial:", err)
		return
	}

	defer rows.Close()

	for rows.Next() {

		var user string
		var msg string

		err := rows.Scan(
			&user,
			&msg,
		)

		if err != nil {
			continue
		}

		ws.WriteMessage(
			1,
			[]byte(user+": "+msg),
		)

	}

}

func handleConnections(
	w http.ResponseWriter,
	r *http.Request,
) {

	ws, err := upgrader.Upgrade(
		w,
		r,
		nil,
	)

	if err != nil {
		return
	}

	defer ws.Close()

	_, p, err := ws.ReadMessage()

	if err != nil {
		return
	}

	username :=
		strings.TrimSpace(
			string(p),
		)

	mutex.Lock()

	clients[username] = ws

	mutex.Unlock()

	fmt.Println(
		"LOG: Usuario conectado:",
		username,
	)

	// Enviar mensajes antiguos

	sendHistory(ws)

	for {

		_, msg, err := ws.ReadMessage()

		if err != nil {

			mutex.Lock()

			delete(
				clients,
				username,
			)

			mutex.Unlock()

			fmt.Println(
				"LOG: Usuario desconectado:",
				username,
			)

			break

		}

		message :=
			string(msg)

		// PRIVADOS

		if strings.HasPrefix(message, "@") {

			parts :=
				strings.SplitN(
					message,
					" ",
					2,
				)

			target :=
				strings.TrimPrefix(
					parts[0],
					"@",
				)

			if len(parts) > 1 {

				private :=
					"(Privado de " +
						username +
						"): " +
						parts[1]

				mutex.Lock()

				if userConn, ok := clients[target]; ok {

					userConn.WriteMessage(
						1,
						[]byte(private),
					)

				}

				mutex.Unlock()

			}

			continue

		}

		// COMUNIDAD

		fullMessage :=
			username +
				": " +
				message

		// Guardar

		saveCommunityMessage(
			username,
			message,
		)

		mutex.Lock()

		for _, client := range clients {

			client.WriteMessage(
				1,
				[]byte(fullMessage),
			)

		}

		mutex.Unlock()

	}

}

func main() {

	var err error

	db, err =
		sql.Open(
			"sqlite3",
			"sway.db",
		)

	if err != nil {
		panic(err)
	}

	_, err =
		db.Exec(`

CREATE TABLE IF NOT EXISTS messages(

id INTEGER PRIMARY KEY AUTOINCREMENT,

username TEXT,

message TEXT,

type TEXT,

created DATETIME DEFAULT CURRENT_TIMESTAMP

)

`)

	if err != nil {
		panic(err)
	}

	fmt.Println(
		"Base de datos SQLite lista",
	)

	http.HandleFunc(
		"/",
		func(
			w http.ResponseWriter,
			r *http.Request,
		) {

			http.ServeFile(
				w,
				r,
				"index.html",
			)

		},
	)

	http.HandleFunc(
		"/ws",
		handleConnections,
	)

	port :=
		os.Getenv("PORT")

	if port == "" {

		port = "8080"

	}

	fmt.Println(
		"Sway activo en puerto:",
		port,
	)

	http.ListenAndServe(
		":"+port,
		nil,
	)

}
