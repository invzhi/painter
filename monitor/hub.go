package monitor

import "github.com/invzhi/shaker/message"

type hub struct {
	// Broadcast send messages to all monitor.
	Broadcast chan *message.Msg

	register   chan *monitor
	unregister chan *monitor
	monitors   map[*monitor]struct{}
}

// NewHub allocates a new hub.
func NewHub() *hub {
	h := &hub{
		Broadcast:  make(chan *message.Msg),
		register:   make(chan *monitor),
		unregister: make(chan *monitor),
		monitors:   make(map[*monitor]struct{}),
	}
	go h.run()
	return h
}

func (h *hub) run() {
	for {
		select {
		case msg := <-h.Broadcast:
			for monitor := range h.monitors {
				monitor.send <- msg
			}
		case monitor := <-h.register:
			h.monitors[monitor] = struct{}{}
		case monitor := <-h.unregister:
			close(monitor.send)
			delete(h.monitors, monitor)
		}
	}
}
