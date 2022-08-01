package broker

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// Example SSE server in Golang.
//     $ go run sse.go
// Inspired from https://gist.github.com/ismasan/3fb75381cd2deb6bfa9c
// Taken (and updated) from https://gist.github.com/Ananto30/8af841f250e89c07e122e2a838698246

type Channel chan Event

type Event struct {
	Event string
	Data  string
}

type Broker struct {
	Name string

	// Events are pushed to this channel by the main events-gathering routine
	Notifier Channel

	// New client connections
	newClients chan Channel

	// Closed client connections
	closingClients chan Channel

	// Client connections registry
	clients map[Channel]struct{}
}

func NewServer(name string) *Broker {
	// Instantiate a broker
	broker := &Broker{
		Name:           name,
		Notifier:       make(Channel, 1),
		newClients:     make(chan Channel),
		closingClients: make(chan Channel),
		clients:        make(map[Channel]struct{}),
	}

	// Set it running - listening and broadcasting events
	go broker.listen()

	return broker
}

type Message interface{}

func (broker *Broker) Stream(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)

	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-store") // originally "no-cache"
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Each connection registers its own message channel with the Broker's connections registry
	messageChan := make(Channel)

	// Signal the broker that we have a new connection
	broker.newClients <- messageChan

	// Remove this client from the map of connected clients
	// when this handler exits.
	defer func() {
		broker.closingClients <- messageChan
	}()

	// Listen to connection close and un-register messageChan
	notify := r.Context().Done()

	for {
		select {
		case <-notify:
			return
		case msg := <-messageChan:
			// Write to the ResponseWriter
			// Server Sent Events compatible
			if msg.Event != "" {
				fmt.Fprintf(w, "event: %s\n", msg.Event)
			}
			fmt.Fprintf(w, "data: %s\n\n", msg.Data)
			// Flush the data immediatly instead of buffering it for later.
			flusher.Flush()
		}
	}

}

func (broker *Broker) listen() {
	for {
		select {
		case s := <-broker.newClients:

			// A new client has connected.
			// Register their message channel
			broker.clients[s] = struct{}{}
			log.Printf("%s: client added. %d registered clients", broker.Name, len(broker.clients))

		case s := <-broker.closingClients:

			// A client has dettached and we want to
			// stop sending them messages.
			delete(broker.clients, s)
			log.Printf("%s: removed client. %d registered clients", broker.Name, len(broker.clients))
			close(s)

		case event := <-broker.Notifier:

			// We got a new event from the outside!
			// Send event to all connected clients
			for clientMessageChan := range broker.clients {
				clientMessageChan <- event
			}
		}
	}

}

func (broker *Broker) Notify(event string, msg Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	broker.Notifier <- Event{
		Event: event,
		Data:  string(data),
	}
	return nil
}
