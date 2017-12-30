package client

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
)

type client struct {
	hub      *hub
	conn     *websocket.Conn
	send     chan []byte
	times    int32
	username string
}

// New allocates a new client and register it to client hub.
func New(h *hub, conn *websocket.Conn, username string) *client {
	c := &client{
		hub:      h,
		conn:     conn,
		send:     make(chan []byte),
		username: username,
	}
	h.Register <- c
	return c
}

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512

	shakeMessage = "shake"
	joinMessage  = "join"
	endMessage   = "end"
)

// ReadTo read message from client and send to channel
func (c *client) ReadTo(send chan []byte) {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(_ string) error {
		log.Print("pong received")
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Print("websocket close going away: ", err)
			}
			break
		}
		if string(message) == shakeMessage {
			c.times++
		}
		// log.Print(c.username, " send: ", string(message))
		send <- []byte(c.username)
	}
}

// Write write message from client's send to conn
func (c *client) Write() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, nil)
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Print("write message error: ", err)
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Print("write ping message error: ", err)
				return
			}
		}
	}
}
