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



func saveMessage(username string, message string, msgType string){

	_, err := db.Exec(
		"INSERT INTO messages(username,message,type) VALUES(?,?,?)",
		username,
		message,
		msgType,
	)


	if err != nil {
		fmt.Println("Error guardando:", err)
	}

}




func sendHistory(ws *websocket.Conn){


	rows, err := db.Query(
		"SELECT username,message FROM messages ORDER BY id ASC",
	)


	if err != nil {
		return
	}


	defer rows.Close()



	for rows.Next(){

		var user string
		var msg string


		rows.Scan(
			&user,
			&msg,
		)



		ws.WriteMessage(
			1,
			[]byte(
				user+": "+msg,
			),
		)

	}

}





func handleConnections(
	w http.ResponseWriter,
	r *http.Request,
){


	ws, err := upgrader.Upgrade(
		w,
		r,
		nil,
	)


	if err != nil {
		return
	}


	defer ws.Close()



	ws.WriteMessage(
		1,
		[]byte(
			"SISTEMA: Escribe tu nombre para entrar",
		),
	)



	_, p, _ := ws.ReadMessage()


	username :=
		strings.TrimSpace(
			string(p),
		)




	mutex.Lock()


	clients[username] = ws



	fmt.Println(
		"Usuario conectado:",
		username,
	)



	// Cargar comunidad

	sendHistory(ws)



	mutex.Unlock()





	for {


		_, msg, err := ws.ReadMessage()


		if err != nil {


			mutex.Lock()


			delete(
				clients,
				username,
			)


			mutex.Unlock()


			break
		}



		message := string(msg)





		// PRIVADOS

		if strings.HasPrefix(message,"@"){


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




			mutex.Lock()



			if conn, ok := clients[target]; ok {



				if len(parts)>1{


					private :=
						"(Privado de "+
						username+
						"): "+
						parts[1]



					conn.WriteMessage(
						1,
						[]byte(private),
					)


				}


			}



			mutex.Unlock()




		}else{



			// COMUNIDAD


			full :=
				username+
				": "+
				message



			// GUARDAR EN SQLITE

			saveMessage(
				username,
				message,
				"community",
			)




			mutex.Lock()



			for _,client := range clients{


				client.WriteMessage(
					1,
					[]byte(full),
				)


			}



			mutex.Unlock()

		}



	}


}






func main(){



	var err error



	db, err =
		sql.Open(
			"sqlite3",
			"sway.db",
		)



	if err != nil{
		panic(err)
	}




	_,err =
		db.Exec(`

		CREATE TABLE IF NOT EXISTS messages(

			id INTEGER PRIMARY KEY AUTOINCREMENT,

			username TEXT,

			message TEXT,

			type TEXT,

			created DATETIME DEFAULT CURRENT_TIMESTAMP

		)

		`)



	if err != nil{
		panic(err)
	}






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





	port :=
		os.Getenv("PORT")



	if port==""{

		port="8080"

	}



	fmt.Println(
		"Sway funcionando en:",
		port,
	)



	http.ListenAndServe(
		":"+port,
		nil,
	)


}
