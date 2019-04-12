package main

import (
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func main() {
	coordinator := &ClientCoordinator{
		clients:     make(map[*Client]bool),
		subscribe:   make(chan *Client),
		unsubscribe: make(chan *Client),
		broadcast:   make(chan []byte),
	}
	go coordinator.run()

	go packetCapture(coordinator.broadcast)

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, _ := upgrader.Upgrade(w, r, nil)
		client := &Client{conn: conn, coordinator: coordinator, send: make(chan []byte, 25600)}
		client.run()
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	http.ListenAndServe(":3000", nil)
}
