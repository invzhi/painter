package client

type hub struct {
	// Broadcast send messages to all client.
	Broadcast chan []byte

	register   chan *client
	unregister chan *client
	Clients    map[string]*client
}

// NewHub allocates a new hub.
func NewHub() *hub {
	h := &hub{
		Broadcast:  make(chan []byte),
		register:   make(chan *client),
		unregister: make(chan *client),
		Clients:    make(map[string]*client),
	}
	go h.run()
	return h
}

func (h *hub) run() {
	for {
		select {
		case message := <-h.Broadcast:
			for _, client := range h.Clients {
				client.send <- message
			}
		case client := <-h.register:
			h.Clients[client.username] = client
		case client := <-h.unregister:
			close(client.send)
			delete(h.Clients, client.username)
		}
	}
}
