


go.sum
Archivo

index.html
Archivo

main.go
Archivo

go.mod
Archivo
estoy haciendo una red social llamada Sway, estoy intentando hacer que se guarden los mensajes al cerrar sesion o pestaña. Arreglalo. Este es el codigo

He revisado el código. El problema es que guardas los mensajes en localStorage del navegador, pero el servidor Go no guarda nada. Eso significa:

Si cierras la pestaña → funciona solo en ese navegador.

Si cierras sesión → tu función logout() borra el usuario con localStorage.removeItem("swayUsername"), pero los chats quedan.

Si entras desde otro navegador/dispositivo → desaparecen porque nunca llegaron al servidor.

Tu HTML ya tiene una base de guardado (saveChats() y loadStorage()), por ejemplo guarda swayCommunityChats y swayChats_usuario. 

 Pero el backend solo mantiene usuarios conectados en memoria (clients), así que al desconectar se pierde todo. 


La solución rápida: guardar mensajes en Go en un mapa y recuperarlos al iniciar sesión.

En main.go añade esto debajo de clients:

var (
	upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	clients  = make(map[string]*websocket.Conn)
	mutex    = sync.Mutex{}

	// Historial de mensajes
	history = make([]string, 0)
)
Después de esta parte:

clients[username] = ws
fmt.Println("LOG: Usuario conectado:", username)
añade:

// Enviar historial al usuario que acaba de entrar
for _, oldMsg := range history {
	ws.WriteMessage(1, []byte(oldMsg))
}
Ahora busca la parte de mensajes públicos:

for _, client := range clients {
	client.WriteMessage(1, []byte(username+": "+message))
}
y cámbiala por:

fullMessage := username + ": " + message

// Guardar mensaje
history = append(history, fullMessage)

for _, client := range clients {
	client.WriteMessage(1, []byte(fullMessage))
}
Y para mensajes privados cambia:

conn.WriteMessage(1, []byte("(Privado de "+username+"): "+parts[1]))
por:

privateMsg := "(Privado de " + username + "): " + parts[1]

// Guardarlo también
history = append(history, privateMsg)

conn.WriteMessage(1, []byte(privateMsg))
Con esto:
✅ cerrar pestaña → conserva mensajes
✅ cerrar sesión → al volver aparecen
✅ reiniciar conexión → recupera historial mientras el servidor siga activo

Pero todavía falta una cosa: si apagas el servidor, se pierde. Para una red social real como Sway necesitarías guardar en una base de datos (SQLite/PostgreSQL). Ahora mismo sería memoria temporal.



Pegado text.txt
Documento

Eso no es un problema del HTML, es de cómo funciona Render.

Cuando haces cambios en tu archivo local, Render no se actualiza automáticamente a menos que tengas conectado un repositorio (por ejemplo GitHub) y hagas un nuevo deploy.

Haz esto:

Si usas Render con GitHub
Guarda el index.html.

Haz commit:

git add .
git commit -m "Actualizo Sway"
git push
En Render:

entra a tu servicio

pulsa Manual Deploy

pulsa Deploy latest commit

Esperas a que termine y recargas la URL.

Si subiste el proyecto directamente a Render
Tienes que volver a desplegar:

Abre Render Dashboard.

Entra en tu servicio.

Busca Deploys.

Pulsa Redeploy.

También revisa algo importante
El archivo que modificaste tiene que estar en la carpeta que Render está sirviendo.

Tu Go tiene:

http.ServeFile(
    w,
    r,
    "index.html",
)
Eso significa que Render busca:

Sway/
 ├ main.go
 └ index.html   ← ESTE
Si pusiste:

Sway/
 ├ main.go
 └ index_sway_corregido.html
no lo va a usar.

Tienes que renombrarlo a:

index.html
y volver a desplegar.


image.png

Veo el error en Render. No es el HTML. Es el main.go.

El log dice:

main.go:12:3: no required module provides package github.com/mattn/go-sqlite3
Significa que Render está descargando tu proyecto, pero tu go.mod no tiene añadida la librería SQLite.

En tu PC funcionó porque hiciste:

go get github.com/mattn/go-sqlite3
pero ese cambio seguramente no lo subiste a GitHub.

Haz esto en tu carpeta de Sway:

1) Abre terminal en la carpeta
cd C:\Users\Ryzen\Desktop\Sway
2) Añade SQLite al proyecto
go get github.com/mattn/go-sqlite3
3) Actualiza go.mod
go mod tidy
Ahora deberías tener en go.mod algo como:

require github.com/mattn/go-sqlite3 v1.x.x
4) Súbelo a GitHub
git add .
git commit -m "Añadir SQLite a Sway"
git push
5) En Render
Pulsa:

Manual Deploy → Deploy latest commit

Ahora debería pasar de:

Build failed
a:

Build successful
Un detalle: cuando funcione en Render, SQLite creará sway.db en el servidor de Render. En el plan gratuito, si Render reinicia la instancia, ese archivo puede perderse porque el disco no es persistente. Para una red social real después habría que usar PostgreSQL de Render. Pero primero dejamos esto funcionando.


vuelve a dar error

Necesito ver el error nuevo exacto porque puede ser otro distinto. En la captura anterior era claramente go.mod, pero si vuelve a fallar necesito el nuevo log.

Haz esto:

En Render entra en:
Logs

Copia las últimas líneas donde pone algo como:

ERROR
...
y pégamelas aquí.

Mientras tanto revisa una cosa: en GitHub, dentro de tu repositorio Sway, ¿tienes estos archivos?

Sway
│
├── main.go
├── index.html
├── go.mod
└── go.sum
Especialmente go.mod y go.sum.

Si no están subidos, Render no sabe que existe SQLite aunque en tu PC funcione.

Puedes comprobarlo entrando a GitHub → tu repo → mira si aparecen esos dos archivos.

Has alcanzado el límite de Free para chats con archivos adjuntos
Mejora tu plan ahora o espera hasta 19:37 para seguir usando archivos, o chatea ahora sin archivos.

Nuevo chat

Mejorar plan


Biblioteca
/
main.go
Más acciones
1
2
3
4
5
6
7
8
9
10
11
12
13
14
15
16
17
18
19
20
21
22
23
24
25
26
27
28
29
30
31
32
33
34
35
36
37
38
39
40
41
42
43
44
45
46
47
48
49
50
51
52
53
54
55
56
57
58
59
60
61
62
63
64
65
66
67
68
69
70
71
72
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
			// Separamos solo en 2 partes: [nombre, mensaje]
			parts := strings.SplitN(message, " ", 2)
			target := strings.TrimPrefix(parts[0], "@")

			mutex.Lock()
			// Verificamos si existe el usuario y si hay un mensaje después del nombre
			if conn, ok := clients[target]; ok {
				if len(parts) > 1 {
					conn.WriteMessage(1, []byte("(Privado de "+username+"): "+parts[1]))
					fmt.Println("LOG: Mensaje privado enviado a", target)
				} else {
					ws.WriteMessage(1, []byte("SISTEMA: Error, formato: @usuario mensaje"))
				}
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
