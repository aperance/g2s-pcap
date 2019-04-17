package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

func main() {
	coordinator := &ClientCoordinator{
		clients:     make(map[*Client]bool),
		subscribe:   make(chan *Client),
		unsubscribe: make(chan *Client),
		broadcast:   make(chan *HttpPacket),
	}

	go coordinator.run()

	go packetCapture(coordinator.broadcast)

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		var upgrader = websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		}
		conn, _ := upgrader.Upgrade(w, r, nil)

		client := &Client{
			conn:        conn,
			coordinator: coordinator,
			send:        make(chan *HttpPacket),
		}

		go client.read()
		go client.write()

		client.coordinator.subscribe <- client

		log.Println("Websocket client connected")
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	http.ListenAndServe(":3000", nil)
}
