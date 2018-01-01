package main

import (
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/websocket"

	"github.com/invzhi/shaker/client"
	"github.com/invzhi/shaker/dashboard"
	"github.com/invzhi/shaker/message"
)

const bufferSize = 256

var (
	globalClientHub    = client.NewHub()
	globalDashboardHub = dashboard.NewHub()

	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

func main() {
	port := ":8080"
	if len(os.Args) == 2 {
		port = os.Args[1]
	}

	http.HandleFunc("/", start)
	http.HandleFunc("/dashboard", monitor)

	http.HandleFunc("/ws", clientWebSocket)
	http.HandleFunc("/dashboardws", dashboardWebSocket)

	log.Fatal(http.ListenAndServe(port, nil))
}

func start(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles("templates/base.html", "templates/start.html"))
	if err := t.Execute(w, nil); err != nil {
		log.Print("start template execute error: ", err)
	}
}

func monitor(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles("templates/base.html", "templates/monitor.html"))
	if err := t.Execute(w, globalClientHub.Clients); err != nil {
		log.Print("monitor template execute error: ", err)
	}
}

func clientWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("client cannot upgrade to websocket: ", err)
		return
	}

	_, msg, err := conn.ReadMessage()
	if err != nil {
		log.Print("someone do not input username: ", err)
		return
	}
	username := string(msg)

	c := client.New(globalClientHub, conn, username)
	globalDashboardHub.Broadcast <- message.New(username, message.Join)

	go c.ReadTo(globalDashboardHub.Broadcast)
	go c.Write()
}

func dashboardWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("monitor cannot upgrade to websocket: ", err)
		return
	}
	d := dashboard.New(globalDashboardHub, conn, bufferSize)

	// update sync

	go d.ReadTo(globalClientHub.Broadcast)
	go d.Write()
}
