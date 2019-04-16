package main

import (
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"
)

type Client struct {
	conn        *websocket.Conn
	coordinator *ClientCoordinator
	send        chan *HttpPacket
}

func (client *Client) run() {
	log.Println("Websocket client connected")
	client.coordinator.subscribe <- client

	for {
		httpPacket := <-client.send

		message, err := json.Marshal(httpPacket)
		if err != nil {
			return
		}

		err = client.conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Println("Websocket client disconnected")
			client.conn.Close()
			client.coordinator.unsubscribe <- client
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
				close(client.send)
			}

		case httpPacket := <-coordinator.broadcast:
			for client := range coordinator.clients {
				client.send <- httpPacket
			}
		}
	}
}
