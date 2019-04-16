package main

import (
	"fmt"
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

	fmt.Print("Enter device name: ")
	var device string
	_, err = fmt.Scanln(&device)

	fmt.Print("Enter port: ")
	var port string
	_, err = fmt.Scanln(&port)

	go packetCapture(device, port, coordinator.broadcast)

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
		client.run()
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	http.ListenAndServe(":3000", nil)
}
