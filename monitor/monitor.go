package monitor

import (
	"log"
	"time"

	"github.com/gorilla/websocket"

	"github.com/invzhi/shaker/message"
)

type monitor struct {
	hub  *hub
	conn *websocket.Conn
	send chan *message.Msg
}

// New allocates a new monitor and register it to monitor hub.
func New(h *hub, conn *websocket.Conn, bufferSize int) *monitor {
	m := &monitor{
		hub:  h,
		conn: conn,
		send: make(chan *message.Msg, bufferSize),
	}
	h.register <- m
	return m
}

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

// ReadTo read message from monitor and send to channel
func (m *monitor) ReadTo(send chan []byte) {
	defer func() {
		m.hub.unregister <- m
		m.conn.Close()
	}()
	pongHandler := func(_ string) error {
		log.Print("pong received.")
		return m.conn.SetReadDeadline(time.Now().Add(pongWait))
	}
	m.conn.SetReadLimit(maxMessageSize)
	m.conn.SetReadDeadline(time.Now().Add(pongWait))
	m.conn.SetPongHandler(pongHandler)

	for {
		_, msg, err := m.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Print("websocket unexpected close error: ", err)
			}
			break
		}
		send <- msg
		log.Print("message to clients: ", string(msg))
	}
}

// Write write message from monitor's send to conn
func (m *monitor) Write() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		m.conn.Close()
	}()
	for {
		select {
		case msg, ok := <-m.send:
			m.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				m.conn.WriteMessage(websocket.CloseMessage, nil)
				return
			}

			if err := m.conn.WriteJSON(msg); err != nil {
				log.Printf("cannot write JSON %s to conn: %v", msg, err)
			}
		case <-ticker.C:
			m.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := m.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Print("cannot write ping message: ", err)
				return
			}
		}
	}
}
