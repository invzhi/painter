package dashboard

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
)

type dashboard struct {
	hub  *hub
	conn *websocket.Conn
	send chan []byte
}

// New allocates a new dashboard and register it to dashboard hub.
func New(h *hub, conn *websocket.Conn, bufferSize int) *dashboard {
	d := &dashboard{
		hub:  h,
		conn: conn,
		send: make(chan []byte, bufferSize),
	}
	h.register <- d
	return d
}

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

// ReadTo read message from dashboard and send to channel
func (d *dashboard) ReadTo(send chan []byte) {
	defer func() {
		d.hub.unregister <- d
		d.conn.Close()
	}()
	d.conn.SetReadLimit(maxMessageSize)
	d.conn.SetReadDeadline(time.Now().Add(pongWait))
	d.conn.SetPongHandler(func(_ string) error {
		log.Print("pong received")
		return d.conn.SetReadDeadline(time.Now().Add(pongWait))
	})
	for {
		_, message, err := d.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Print("websocket close going away: ", err)
			}
			break
		}
		send <- message
		log.Print("message to clients: ", string(message))
	}
}

// Write write message from dashboard's send to conn
func (d *dashboard) Write() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		d.conn.Close()
	}()
	for {
		select {
		case message, ok := <-d.send:
			d.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				d.conn.WriteMessage(websocket.CloseMessage, nil)
				return
			}

			d.conn.WriteMessage(websocket.TextMessage, message)

			w, err := d.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				log.Print("could not get next writer: ", err)
				return
			}
			w.Write(message)

			if err = w.Close(); err != nil {
				log.Print("writer close error: ", err)
				return
			}
		case <-ticker.C:
			d.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := d.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Print("write ping message error: ", err)
				return
			}
		}
	}
}
