package main

import (
	"log"
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
		broadcast:   make(chan *HttpPacket),
	}
	go coordinator.run()
	go packetCapture(coordinator.broadcast)

	log.Println("Capturing HTTP packets...")

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, _ := upgrader.Upgrade(w, r, nil)
		client := &Client{
			conn:        conn,
			coordinator: coordinator,
			send:        make(chan *HttpPacket),
		}
		client.run()
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	http.ListenAndServe(":3000", nil)
}
