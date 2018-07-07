package wsevent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"strings"
	"time"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 1024 * 4
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// eventClient is a middleman between the websocket connection and the hub.
type eventClient struct {
	hub *eventHub

	// The websocket connection.
	conn *websocket.Conn

	// connection time
	connTime time.Time

	// Buffered channel of outbound messages.
	send chan []byte

	// registered
	registered bool

	// args that need to filter event
	registerArgs interface{}
}

func newEventClient(conn *websocket.Conn) *eventClient {
	client := &eventClient{hub: _eventhub, conn: conn, send: make(chan []byte, 512), registered: false, connTime: time.Now()}
	client.hub.registerClient <- client

	time.AfterFunc(time.Duration(_configArgs.RegisterTimeout)*time.Second, func() {
		if !client.registered {
			log.Printf("[wsevent] client %+v is not registered in %d seconds, disconnected it", client, _configArgs.RegisterTimeout)
			client.conn.Close()
		}
	})

	return client
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *eventClient) readPump() {
	defer func() {
		c.hub.unregisterClient <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[wsevent] ReadMessage error: %v", err)
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))

		var param struct {
			Type string      `json:"type"`
			Args interface{} `json:"args"`
		}

		log.Printf("[wsevent] received from client: %s", string(message))

		var respErr error
		var respMsg interface{}

		if err := json.Unmarshal(message, &param); err == nil {
			log.Printf("[wsevent] received from client, param: %+v", param)

			if param.Type == "reg" {
				if args, e := _configArgs.ValidateRegisterArgs(param.Args); e == nil {
					c.registerArgs = args
					respMsg = newResponseMessage(param.Type, newOKResponseData("args", c.registerArgs))
					c.registered = true
				} else {
					respErr = NewError(ErrorUnregistered, e.Error())
					c.registered = false
				}

			} else {
				respErr = NewError(ErrorCodeNotSupported, fmt.Sprintf("Invalid type: %s", param.Type))
			}
		} else {
			respErr = err
		}

		log.Printf("%+v", respErr)

		var respBody []byte
		if respErr != nil {
			respBody, _ = json.Marshal(newResponseMessage(param.Type, newErrorResponseData(respErr)))
		} else {
			respBody, _ = json.Marshal(respMsg)
		}

		c.send <- respBody

	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *eventClient) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

type wsHandler struct {
	path string
}

func (h *wsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)

	if strings.Compare(r.URL.Path, h.path) != 0 {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, fmt.Sprintf("%v", err), http.StatusMethodNotAllowed)
		return
	}

	client := newEventClient(conn)

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}
