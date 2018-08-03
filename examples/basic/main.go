package main

import (
	"github.com/hongjundu/go-websocket-event-hub"
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
