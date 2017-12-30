package main

import (
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/websocket"

	"github.com/invzhi/shaker/client"
	"github.com/invzhi/shaker/dashboard"
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

	staticFileServer := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", staticFileServer))
	// http.HanleFunc("/favio.ico", func(w http.ResponseWriter, r *http.Request) {
	// 	http.ServeFile(w, r, "static/favic")
	// })

	http.HandleFunc("/", index)
	http.HandleFunc("/start", start)
	http.HandleFunc("/dashboard", monitor)

	http.HandleFunc("/startws", clientWebSocket)
	http.HandleFunc("/dashboardws", dashboardWebSocket)

	log.Fatal(http.ListenAndServe(port, nil))
}

func index(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		t := template.Must(template.ParseFiles("templates/base.html", "templates/index.html"))
		if err := t.Execute(w, nil); err != nil {
			log.Print("index template execute error: ", err)
		}
	case http.MethodPost:
		username := template.HTMLEscapeString(r.FormValue("username"))
		http.SetCookie(w, &http.Cookie{Name: "username", Value: username})
		log.Print("set username cookie: ", username)
		http.Redirect(w, r, "/start", http.StatusSeeOther)
	}
}

func start(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles("templates/base.html", "templates/start.html"))
	if err := t.Execute(w, nil); err != nil {
		log.Print("start template execute error: ", err)
	}
}

func monitor(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles("templates/base.html", "templates/dashboard.html"))
	if err := t.Execute(w, nil); err != nil {
		log.Print("dashboard template execute error: ", err)
	}
}

func clientWebSocket(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("username")
	if err != nil {
		log.Print("cookie not found: ", err)
		return
	}

	log.Print(cookie)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("client could not upgrade to websocket: ", err)
		return
	}

	c := client.New(globalClientHub, conn, cookie.Value)
	globalDashboardHub.Broadcast <- []byte("join") // joinMessage

	go c.ReadTo(globalDashboardHub.Broadcast)
	go c.Write()
}

func dashboardWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("dashboard could not upgrade to websocket: ", err)
		return
	}
	d := dashboard.New(globalDashboardHub, conn, bufferSize)

	go d.ReadTo(globalClientHub.Broadcast)
	go d.Write()
}
