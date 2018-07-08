package wsevent

import (
	"fmt"
	"log"
	"net/http"
)

// public APIs

type Options struct {
	EventQueueSize       int
	PublishRoutineNum    int
	LogEventEnabled      bool
	RegisterTimeout      int
	ValidateRegisterArgs func(args interface{}) (interface{}, error)
	FilterEvent          func(args interface{}, event interface{}) bool
}

func Init(path string, options Options) {
	if _eventhub != nil {
		log.Fatalf("[wsevent] already initialized")
	}
	setOptions(options)

	go newEventHub().run()
	handler := &wsHandler{path}

	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		handler.ServeHTTP(w, r)
	})
}

func InitWithPort(path string, port int, options Options) {
	if _eventhub != nil {
		log.Fatalf("[wsevent] already initialized")
	}

	setOptions(options)

	go newEventHub().run()
	handler := &wsHandler{path}

	go func() {
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), handler))
	}()
}

func PublishEvent(event interface{}) {
	if _eventhub == nil {
		log.Fatalf("[wsevent] was not initialized")
	}

	_eventhub.broadcastEvents <- event
}
