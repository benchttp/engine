package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	port  = 8080
	token = "6db67fafc4f5bf965a5a"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return r.URL.Query().Get("access_token") == token
	},
}

func main() {
	addr := fmt.Sprintf("localhost:%d", port)

	fmt.Printf("listening on http://%s\n", addr)

	http.HandleFunc("/run", run)
	http.HandleFunc("/status", status)

	log.Fatal(http.ListenAndServe(addr, nil))
}

func run(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("connected /run")

	defer ws.Close()

	dr := dummyRun{}

	for {
		m, err := readTextMessage(ws)
		if err != nil {
			log.Println("read error:", err)
			break
		}

		switch string(m) {
		case "run":
			if dr.started {
				log.Println("already started")

				err = ws.WriteMessage(websocket.TextMessage, []byte("already started"))
				if err != nil {
					log.Println("cannot write message:", err)
					break
				}
			} else {
				log.Println("starting run")

				dr.run()

				err = ws.WriteMessage(websocket.TextMessage, []byte("ack"))
				if err != nil {
					log.Println("cannot write message:", err)
					break
				}
			}

		default:
			log.Printf("<- %s", m)

			err = ws.WriteMessage(websocket.TextMessage, m)
			if err != nil {
				log.Println("cannot write message:", err)
				break
			}
		}
	}
}

func readTextMessage(ws *websocket.Conn) ([]byte, error) {
	messageType, m, err := ws.ReadMessage()
	if err != nil {
		return nil, fmt.Errorf("cannot read message as text: %s", err)
	}

	if messageType != websocket.TextMessage {
		return nil, fmt.Errorf("cannot read message as text: message type is not text")
	}

	return m, nil
}

type dummyRun struct {
	started bool
}

func (d *dummyRun) run() {
	d.started = true
}

func status(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("connected /status")

	defer ws.Close()

	ds := dummyStatus{}

	for range time.Tick(time.Second * 1) {
		ds.inc()
		data := []byte(fmt.Sprint(ds.c))

		err = ws.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			log.Println("cannot write message:", err)
			break
		}
		log.Printf("-> %s", data)
	}
}

type dummyStatus struct {
	c int
}

func (d *dummyStatus) inc() {
	d.c++
}
