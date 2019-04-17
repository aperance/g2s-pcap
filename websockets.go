package main

import (
	"log"

	"github.com/gorilla/websocket"
)

type Client struct {
	conn        *websocket.Conn
	coordinator *ClientCoordinator
	send        chan *HttpPacket
}

func (client *Client) read() {
	defer func() {
		client.coordinator.unsubscribe <- client
		close(client.send)
		client.conn.Close()
	}()

	for {
		_, _, err := client.conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
	}
}

func (client *Client) write() {
	for {
		err = client.conn.WriteJSON(<-client.send)
		if err != nil {
			log.Println(err)
			return
		}
	}
}

type ClientCoordinator struct {
	clients     map[*Client]bool
	subscribe   chan *Client
	unsubscribe chan *Client
	broadcast   chan *HttpPacket
}

func (coordinator *ClientCoordinator) run() {
	for {
		select {

		case client := <-coordinator.subscribe:
			coordinator.clients[client] = true

		case client := <-coordinator.unsubscribe:
			if _, ok := coordinator.clients[client]; ok {
				delete(coordinator.clients, client)
			}

		case httpPacket := <-coordinator.broadcast:
			for client := range coordinator.clients {
				client.send <- httpPacket
			}
		}
	}
}
