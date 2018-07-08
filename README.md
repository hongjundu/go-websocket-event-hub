# go-websocket-event-hub

Golang event push library. Makes web push notification easy via websocket. dependency: [github.com/gorilla/websocket](github.com/gorilla/websocket)

## Public API

* Init event hub with one of the following APIs

        func Init(path string, options Options) 
        func InitWithPort(path string, port int, options Options) 
        
    ```Init(path string, options Options) ``` initializes the event hub and listen on default http handler.
           
        wsevent.Init("/wsevents", wsevent.Options{})
        log.Fatal(http.ListenAndServe(":8080", nil))
            
    ```InitWithPort(path string, port int, options Options) ```initializes the event hub and listen on a given port.
        
        wsevent.InitWithPort("/wsevents", 8081, wsevent.Options{}))
        
    ```Options``` argument in Init function
    The default options ```wsevent.Options{}``` just works. If you want to customize for your needs, following optons are avaiable.
    
* ```EventQueueSize int```
* ```PublishRoutineNum int```
* ```LogEventEnabled bool```
* ```ValidateRegisterArgs func(args interface{}) (interface{}, error)```
* ```FilterEvene func(args interface{}, event interface{}) bool```
    
* Publish event to each registered web socket clients with following API
    
        func PublishEvent(event interface{})

## Basic example

    package main

    import (
        "github.com/du-hj/go-websocket-event-hub"
        "log"
        "net/http"
        "time"
    )

    func homePage(w http.ResponseWriter, r *http.Request) {
        log.Println(r.URL)

        if r.URL.Path != "/" {
            http.Error(w, "Not found", http.StatusNotFound)
            return
        }
        if r.Method != "GET" {
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
            return
        }
        http.ServeFile(w, r, "index.html")
    }

    func main() {
        log.Printf("server starts...")

        //wsevent.InitWithPort("/wsevents", 8081, wsevent.Options{}))
        wsevent.Init("/wsevents", wsevent.Options{})

        publishEvents()

        http.HandleFunc("/", homePage)
        log.Fatal(http.ListenAndServe(":8080", nil))

        log.Printf("server exits")
    }

    type event struct {
        Event string `json:"event"`
        From  int    `json:"from"`
    }

    func publishEvents() {
        for i := 0; i < 10; i++ {
            index := i + 1
            go func() {
                for {
                    wsevent.PublishEvent(&event{Event: "test", From: index})
                    time.Sleep(time.Second * 5)
                }
            }()
        }
    }


## Advanced example

    package main

    import (
        "encoding/json"
        "github.com/du-hj/go-websocket-event-hub"
        "log"
        "net/http"
        "time"
    )

    func homePage(w http.ResponseWriter, r *http.Request) {
        log.Println(r.URL)

        if r.URL.Path != "/" {
            http.Error(w, "Not found", http.StatusNotFound)
            return
        }
        if r.Method != "GET" {
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
            return
        }
        http.ServeFile(w, r, "index.html")
    }

    func main() {
        log.Printf("server starts...")

        options := wsevent.Options{
            EventQueueSize:       10240,
            PublishRoutineNum:    8,
            LogEventEnabled:      false,
            RegisterTimeout:      5,
            ValidateRegisterArgs: validateRegisterArgs,
            FilterEvent:          filterEvent}

        wsevent.Init("/wsevents", options)
        //wsevent.InitWithPort("/wsevents", 8081, options)

        publishEvents()

        http.HandleFunc("/", homePage)
        log.Fatal(http.ListenAndServe(":8080", nil))

        log.Printf("server exists")
    }

    type event struct {
        Event string `json:"event"`
        From  int    `json:"from"`
    }

    type registerArgs struct {
        Token string `json:"token"`
        Hint  string `json:"hint"`
    }

    func validateRegisterArgs(args interface{}) (typedArgs interface{}, err error) {
        log.Printf("validateRegisterArgs: %+v", args)

        if args == nil {
            err = wsevent.NewError(wsevent.ErrorUnregistered, "No register args")
            return
        }

        body, e := json.Marshal(args)
        if e != nil {
            err = wsevent.NewError(wsevent.ErrorUnregistered, "Invalid register args fromat")
            return
        }

        var regArgs registerArgs

        e = json.Unmarshal(body, &regArgs)
        if e != nil {
            err = wsevent.NewError(wsevent.ErrorUnregistered, "Invalid register args format")
            return
        }

        if len(regArgs.Token) == 0 {
            err = wsevent.NewError("unauthorized", "Invalid register args: no token present")
            return
        }

        // verify token in real project
        if regArgs.Token != "123" {
            err = wsevent.NewError("unauthorized", "Invalid register args: wrong token")
            return
        }

        typedArgs = &regArgs
        return
    }

    func filterEvent(args interface{}, evt interface{}) bool {
        log.Printf("Filter Event: args: %+v event: %+v", args, evt)

        if evt == nil {
            log.Printf("FilterEvent: event is nil")
            return false
        }

        if args == nil {
            log.Printf("FilterEvent: args is nil")
            return true
        }

        var typedArgs *registerArgs
        var ok bool

        if typedArgs, ok = args.(*registerArgs); !ok {
            log.Printf("FilterEvent: invlid args type")
            return false
        }

        var typedEvent *event
        if typedEvent, ok = evt.(*event); !ok {
            log.Printf("FilterEvent: invlid event type")
            return false
        }

        // No filter
        if len(typedArgs.Hint) == 0 {
            return true
        }

        if typedArgs.Hint == "odd" {
            return typedEvent.From%2 != 0
        } else if typedArgs.Hint == "even" {
            return typedEvent.From%2 == 0
        }

        return true
    }

    func publishEvents() {
        for i := 0; i < 10; i++ {
            index := i + 1
            go func() {
                for {
                    wsevent.PublishEvent(&event{Event: "test", From: index})
                    time.Sleep(time.Second * 1)
                }
            }()
        }
    }