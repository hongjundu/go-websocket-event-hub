package wsevent

import (
	"fmt"
	"log"
	"net/http"
)

// public APIs

type ConfigArgs struct {
	EventQueueSize       int
	PublishRoutineNum    int
	LogEventEnabled      bool
	RegisterTimeout      int
	ValidateRegisterArgs func(args interface{}) (interface{}, error)
	FilterEvent          func(args interface{}, event interface{}) bool
}

func Config(args ConfigArgs) {
	if args.EventQueueSize > 0 {
		_configArgs.EventQueueSize = args.EventQueueSize
	}

	if args.PublishRoutineNum > 0 {
		_configArgs.PublishRoutineNum = args.PublishRoutineNum
	}

	if args.RegisterTimeout > 0 {
		_configArgs.RegisterTimeout = args.RegisterTimeout
	}

	if args.ValidateRegisterArgs != nil {
		_configArgs.ValidateRegisterArgs = args.ValidateRegisterArgs
	}

	if args.FilterEvent != nil {
		_configArgs.FilterEvent = args.FilterEvent
	}

	_configArgs.LogEventEnabled = args.LogEventEnabled
}

func PublishEvent(event interface{}) {
	if _eventhub == nil {
		log.Fatalf("[wsevent] not initialized")
	}

	_eventhub.broadcastEvents <- event
}

func InitWithPort(path string, port int) {
	if _eventhub != nil {
		log.Fatalf("[wsevent] already initialized")
	}

	go newEventHub().run()
	handler := &wsHandler{path}

	go func() {
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), handler))
	}()
}

func Init(path string) {
	if _eventhub != nil {
		log.Fatalf("[wsevent] already initialized")
	}

	go newEventHub().run()
	handler := &wsHandler{path}

	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		handler.ServeHTTP(w, r)
	})
}
