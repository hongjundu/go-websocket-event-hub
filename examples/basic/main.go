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

	//wsevent.InitWithPort("/wsevents", 8081)

	wsevent.Init("/wsevents")

	publishExampleEvents()

	http.HandleFunc("/", homePage)
	log.Fatal(http.ListenAndServe(":8080", nil))

	log.Printf("server exits")
}

type event struct {
	Event string    `json:"event"`
	From  int       `json:"from"`
	Time  time.Time `json:"time"`
}

func publishExampleEvents() {
	for i := 0; i < 10; i++ {
		index := i + 1
		go func() {
			for {
				wsevent.PublishEvent(&event{Event: "test", From: index, Time: time.Now()})
				time.Sleep(time.Second * 1)
			}
		}()
	}
}
