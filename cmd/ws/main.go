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

	http.HandleFunc("/run", handleRun)

	log.Fatal(http.ListenAndServe(addr, nil))
}

type Runner struct {
	running  bool
	progress progress

	isPollingProgress    bool
	abortPollingProgress chan struct{}
}

func (r *Runner) run() string {
	if !r.running {
		r.running = true
		r.progress = progress{}

		// Prepare abort signal.
		r.abortPollingProgress = make(chan struct{}, 1)

		return "running"
	}

	return "error: already running"
}

func (r *Runner) stop() string {
	if !r.running {
		return "error: not running"
	}

	r.running = false
	r.abortPollingProgress <- struct{}{}
	r.isPollingProgress = false
	r.progress = progress{}

	return "stopped"
}

type progress struct {
	value int
}

func (p *progress) next() int {
	// Go no further than 100.
	if p.value <= 100 {
		p.value++
	}

	return p.value
}

func handleRun(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	log.Printf("websocket connected with client %s", r.Host)

	defer ws.Close()

	runner := Runner{}

	for {
		m, err := readTextMessage(ws)
		if err != nil {
			log.Println("read error:", err)
			break
		}

		log.Printf("<- %s", m)

		switch string(m) {
		case "run":
			status := runner.run()

			err = writeTextMessage(ws, status)
			if err != nil {
				break
			}

		case "stop":
			status := runner.stop()

			err = writeTextMessage(ws, status)
			if err != nil {
				break
			}

		case "pull":
			err = writeTextMessage(ws, "not implemented")
			if err != nil {
				break
			}
			log.Println("^^ not implemented", m)
		default:
			log.Println("^^ not implemented", m)
		}

		if runner.running && !runner.isPollingProgress {
			go runner.pollProgress(ws)
		}
	}
}

func (r *Runner) pollProgress(ws *websocket.Conn) {
	r.isPollingProgress = true

	for range time.Tick(time.Millisecond * 500) {
		select {
		case <-r.abortPollingProgress:
			return
		default:
		}

		if r.progress.value == 100 {
			r.isPollingProgress = false
			return
		}

		val := r.progress.next()
		percent := fmt.Sprintf("%d%%", val)

		err := writeTextMessage(ws, percent)
		if err != nil {
			break
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

func writeTextMessage(ws *websocket.Conn, m string) error {
	err := ws.WriteMessage(websocket.TextMessage, []byte(m))
	if err != nil {
		log.Println("cannot write message:", err)
		return err
	}

	log.Printf("-> %s", m)

	return nil
}
