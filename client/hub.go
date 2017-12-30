package client

type hub struct {
	Broadcast  chan []byte
	Register   chan *client
	unregister chan *client
	clients    map[string]*client
}

// NewHub allocates a new hub.
func NewHub() *hub {
	h := &hub{
		Broadcast:  make(chan []byte),
		Register:   make(chan *client),
		unregister: make(chan *client),
		clients:    make(map[string]*client),
	}
	go func() {
		for {
			select {
			case client := <-h.Register:
				h.clients[client.username] = client
			case client := <-h.unregister:
				close(client.send)
				delete(h.clients, client.username)
			case message := <-h.Broadcast:
				for _, client := range h.clients {
					client.send <- message
				}
			}
		}
	}()
	return h
}
