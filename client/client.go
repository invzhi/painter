package client

import (
	"log"
	"time"

	"github.com/gorilla/websocket"

	"github.com/invzhi/shaker/message"
)

type client struct {
	hub      *hub
	conn     *websocket.Conn
	send     chan []byte
	username string

	// Times is shake score
	Times int32
}

// New allocates a new client and register it to client hub.
func New(h *hub, conn *websocket.Conn, username string) *client {
	c := &client{
		hub:      h,
		conn:     conn,
		send:     make(chan []byte),
		username: username,
	}
	h.register <- c
	return c
}

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

// ReadTo read message from client and send to channel
func (c *client) ReadTo(send chan *message.Msg) {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	pongHandler := func(_ string) error {
		log.Print("pong received.")
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	}
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(pongHandler)

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Print("websocket unexpected close error: ", err)
			}
			break
		}
		c.Times++
		send <- message.New(c.username, message.Shake)
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
		case msg, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, nil)
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				log.Printf("cannot write text message '%s': %v", msg, err)
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Print("cannot write ping message: ", err)
				return
			}
		}
	}
}
