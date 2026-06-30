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
	clients = make(map[string]*websocket.Conn)
	mutex = sync.Mutex{}
)
		},
	}

	clients = make(map[string]*websocket.Conn)

	mutex = sync.Mutex{}

	// HISTORIAL DE MENSAJES
	history = make([]string, 0)
)


func handleConnections(w http.ResponseWriter, r *http.Request) {

	ws, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		return
	}

	defer ws.Close()


	// Entrada usuario
	ws.WriteMessage(1, []byte("SISTEMA: Escribe tu nombre para entrar"))


	_, p, _ := ws.ReadMessage()


	username := strings.TrimSpace(string(p))


	mutex.Lock()


	clients[username] = ws


	fmt.Println("LOG: Usuario conectado:", username)



	// Enviar historial al entrar
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



		fmt.Println(
			"LOG:",
			username,
			":",
			message,
		)




		// MENSAJES PRIVADOS

		if strings.HasPrefix(message, "@") {


			parts := strings.SplitN(message, " ", 2)



			target := strings.TrimPrefix(parts[0], "@")



			mutex.Lock()



			if conn, ok := clients[target]; ok {


				if len(parts) > 1 {



					privateMessage :=
						"(Privado de " +
						username +
						"): " +
						parts[1]



					// Guardar privado

					history = append(
						history,
						privateMessage,
					)



					conn.WriteMessage(
						1,
						[]byte(privateMessage),
					)



					fmt.Println(
						"LOG: Privado enviado a",
						target,
					)



				} else {


					ws.WriteMessage(
						1,
						[]byte(
							"SISTEMA: Error formato @usuario mensaje",
						),
					)

				}



			} else {


				ws.WriteMessage(
					1,
					[]byte(
						"SISTEMA: Usuario "+target+" no encontrado",
					),
				)

			}



			mutex.Unlock()



		} else {



			// MENSAJE COMUNIDAD


			fullMessage :=
				username +
				": " +
				message



			mutex.Lock()



			// Guardar mensaje

			history = append(
				history,
				fullMessage,
			)



			// Enviar a todos

			for _, client := range clients {


				client.WriteMessage(
					1,
					[]byte(fullMessage),
				)

			}



			mutex.Unlock()

		}


	}

}




func main(){


	http.HandleFunc(
		"/",
		func(
			w http.ResponseWriter,
			r *http.Request,
		){

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



	port := os.Getenv("PORT")



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
