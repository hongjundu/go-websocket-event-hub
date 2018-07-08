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
	registerClient chan *eventClient

	// Unregister requests from clients.
	unregisterClient chan *eventClient
}

var _eventhub *eventHub

func newEventHub() *eventHub {
	_eventhub = &eventHub{
		broadcastEvents:  make(chan interface{}, _globalOptions.EventQueueSize),
		registerClient:   make(chan *eventClient),
		unregisterClient: make(chan *eventClient),
		clients:          make(map[*eventClient]bool),
	}
	return _eventhub
}

func (h *eventHub) run() {
	for i := 0; i < _globalOptions.PublishRoutineNum; i++ {
		go func() {
			for event := range h.broadcastEvents {
				if _globalOptions.LogEventEnabled {
					log.Printf("[wsevent] connectd: %d publish event: %+v ", len(h.clients), event)
				}

				for client := range h.clients {
					if !client.registered {
						log.Printf("[wsevent] client is not registered: %+v", client)
						continue
					}

					if _globalOptions.FilterEvent(client.registerArgs, event) {
						if message, err := json.Marshal(newResponseMessage("event", event)); err == nil {
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
		case client := <-h.registerClient:
			log.Printf("[wsevent] client open: %+v", client)
			h.clients[client] = true
		case client := <-h.unregisterClient:
			log.Printf("[wsevent] client close: %+v", client)
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		}
	}

}

var _globalOptions Options = Options{EventQueueSize: 1024, PublishRoutineNum: 4, RegisterTimeout: 60}

func setOptions(options Options) {
	if options.EventQueueSize > 0 {
		_globalOptions.EventQueueSize = options.EventQueueSize
	}

	if options.PublishRoutineNum > 0 {
		_globalOptions.PublishRoutineNum = options.PublishRoutineNum
	}

	if options.RegisterTimeout > 0 {
		_globalOptions.RegisterTimeout = options.RegisterTimeout
	}

	if options.ValidateRegisterArgs != nil {
		_globalOptions.ValidateRegisterArgs = options.ValidateRegisterArgs
	}

	if options.FilterEvent != nil {
		_globalOptions.FilterEvent = options.FilterEvent
	}

	_globalOptions.LogEventEnabled = options.LogEventEnabled
}

func init() {
	_globalOptions.ValidateRegisterArgs = func(args interface{}) (interface{}, error) {
		return nil, nil
	}
	_globalOptions.FilterEvent = func(args interface{}, event interface{}) bool {
		return true
	}
}
