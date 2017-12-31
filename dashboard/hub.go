package dashboard

type hub struct {
	// Broadcast send messages to all dashboard.
	Broadcast chan []byte

	register   chan *dashboard
	unregister chan *dashboard
	dashboards map[*dashboard]struct{}
}

// NewHub allocates a new hub.
func NewHub() *hub {
	h := &hub{
		Broadcast:  make(chan []byte),
		register:   make(chan *dashboard),
		unregister: make(chan *dashboard),
		dashboards: make(map[*dashboard]struct{}),
	}
	go h.run()
	return h
}

func (h *hub) run() {
	for {
		select {
		case message := <-h.Broadcast:
			for dashboard := range h.dashboards {
				dashboard.send <- message
			}
		case dashboard := <-h.register:
			h.dashboards[dashboard] = struct{}{}
		case dashboard := <-h.unregister:
			close(dashboard.send)
			delete(h.dashboards, dashboard)
		}
	}
}
