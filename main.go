package main

import (
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/websocket"

	"github.com/invzhi/shaker/client"
	"github.com/invzhi/shaker/message"
	"github.com/invzhi/shaker/monitor"
)

const bufferSize = 256

var (
	globalClientHub  = client.NewHub()
	globalMonitorHub = monitor.NewHub()

	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

func init() {
	staticFileServer := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", staticFileServer))

	http.HandleFunc("/", start)
	http.HandleFunc("/monitor", monitoring)

	http.HandleFunc("/ws", clientWebSocket)
	http.HandleFunc("/monitorws", monitorWebSocket)
}

func main() {
	port := ":8080"
	if len(os.Args) == 2 {
		port = os.Args[1]
	}

	log.Fatal(http.ListenAndServe(port, nil))
}

func start(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles("templates/base.html", "templates/start.html"))
	if err := t.Execute(w, nil); err != nil {
		log.Print("start template execute error: ", err)
	}
}

func monitoring(w http.ResponseWriter, r *http.Request) {
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
	globalMonitorHub.Broadcast <- message.New(username, message.Join)

	go c.ReadTo(globalMonitorHub.Broadcast)
	go c.Write()
}

func monitorWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("monitor cannot upgrade to websocket: ", err)
		return
	}
	m := monitor.New(globalMonitorHub, conn, bufferSize)

	// update sync

	go m.ReadTo(globalClientHub.Broadcast)
	go m.Write()
}
