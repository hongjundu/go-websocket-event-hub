package wsevent

import (
	"encoding/json"
	"log"
)

// eventHub maintains the set of active clients and broadcasts messages to the clients
type eventHub struct {
	// Registered clients.
	clients map[*eventClient]bool

	// events to be broadcast to clients
	broadcastEvents chan interface{}

	// Register requests from the clients.
	register chan *eventClient

	// Unregister requests from clients.
	unregister chan *eventClient
}

var _eventhub *eventHub

func newEventHub() *eventHub {
	_eventhub = &eventHub{
		broadcastEvents: make(chan interface{}, _configArgs.EventQueueSize),
		register:        make(chan *eventClient),
		unregister:      make(chan *eventClient),
		clients:         make(map[*eventClient]bool),
	}
	return _eventhub
}

func (h *eventHub) run() {
	for i := 0; i < _configArgs.PublishRoutineNum; i++ {
		go func() {
			for event := range h.broadcastEvents {
				if _configArgs.LogEvent {
					log.Printf("[wsevent] connectd: %d publish event: %+v ", len(h.clients), event)
				}

				for client := range h.clients {
					if !client.authorized {
						log.Printf("[wsevent] client is not authorized: %+v", client)
						continue
					}

					if _configArgs.FilterEvent(client.registerArgs, event) {
						if message, err := json.Marshal(event); err == nil {
							select {
							case client.send <- message:
							default:
								log.Printf("[wsevent] client seems disconnected: %+v", client)
								close(client.send)
								delete(h.clients, client)
							}
						} else {
							log.Printf("[wsevent] marshal event failed, event: %+v error: %+v", event, err)
						}
					}
				}
			}
		}()
	}

	for {
		select {
		case client := <-h.register:
			log.Printf("[wsevent] client open: %+v", client)
			h.clients[client] = true
		case client := <-h.unregister:
			log.Printf("[wsevent] client close: %+v", client)
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		}
	}

}

var _configArgs ConfigArgs = ConfigArgs{EventQueueSize: 1024, PublishRoutineNum: 4}

func init() {
	_configArgs.ValidateRegisterArgs = func(args interface{}) (interface{}, error) {
		return nil, nil
	}
	_configArgs.FilterEvent = func(args interface{}, event interface{}) bool {
		return true
	}
}
