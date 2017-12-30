package dashboard

type hub struct {
	Broadcast  chan []byte
	Register   chan *dashboard
	unregister chan *dashboard
	dashboards map[*dashboard]struct{}
}

// NewHub allocates a new hub.
func NewHub() *hub {
	h := &hub{
		Broadcast:  make(chan []byte),
		Register:   make(chan *dashboard),
		unregister: make(chan *dashboard),
		dashboards: make(map[*dashboard]struct{}),
	}
	go func() {
		for {
			select {
			case dashboard := <-h.Register:
				h.dashboards[dashboard] = struct{}{}
			case dashboard := <-h.unregister:
				close(dashboard.send)
				delete(h.dashboards, dashboard)
			case message := <-h.Broadcast:
				for dashboard := range h.dashboards {
					dashboard.send <- message
				}
			}
		}
	}()
	return h
}
