package pkg

import (
	"log"
	"sync"

	"github.com/gofiber/contrib/websocket"
)

type EventBody struct {
	Event    string `json:"event"`
	ToPage   string `json:"to_page,omitempty"`
	PageSize string `json:"page_size,omitempty"`
	By       string `json:"by,omitempty"`
}

type TableState struct {
	Page           int
	PageSize       int
	OrderColumn    string
	OrderDirection string
}

func (ts *TableState) UpdateDirection(newColumn string) {
	if ts.OrderColumn == newColumn {
		if ts.OrderDirection == "ASC" {
			ts.OrderDirection = "DESC"
		} else {
			ts.OrderDirection = "ASC"
		}
	} else {
		ts.OrderDirection = "DESC"
	}

	ts.OrderColumn = newColumn
}

func defaultTableState() *TableState {
	return &TableState{Page: 1, PageSize: 10, OrderColumn: "id", OrderDirection: "DESC"}
}

type client struct {
	isClosing bool
	mu        sync.Mutex
}

var clients = make(map[*websocket.Conn]*client)
var states = make(map[*websocket.Conn]*TableState)
var register = make(chan *websocket.Conn)
var broadcast = make(chan string)
var unregister = make(chan *websocket.Conn)

func (c *Container) RunHub() {
	for {
		select {
		case connection := <-register:
			clients[connection] = &client{}
			states[connection] = defaultTableState()

		case message := <-broadcast:
			// Send the message to all clients
			for connection, c := range clients {
				go func(connection *websocket.Conn, c *client) { // send to each client in parallel so we don't block on a slow client
					c.mu.Lock()
					defer c.mu.Unlock()

					if c.isClosing {
						return
					}

					if err := connection.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
						c.isClosing = true
						log.Println("write error:", err)

						connection.WriteMessage(websocket.CloseMessage, []byte{})
						connection.Close()
						unregister <- connection
					}
				}(connection, c)
			}

		case connection := <-unregister:
			// Remove the client from the hub
			delete(clients, connection)
			delete(states, connection)
		}
	}
}
