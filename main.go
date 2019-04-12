package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type HttpPacket struct {
	Source      string `json:"src"`
	Destination string `json:"dst"`
	Protocol    string `json:"type"`
	Payload     string `json:"payload"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func main() {
	receivedPackets := make(chan HttpPacket)

	go packetCapture(receivedPackets)
	fmt.Println("Capturing HTTP packets...")

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, _ := upgrader.Upgrade(w, r, nil)

		go writeMessage(conn, receivedPackets)

		_, _, err := conn.ReadMessage()
		if err != nil {
			fmt.Printf("error: %v\n", err)
			return
		}
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	http.ListenAndServe(":3000", nil)
}

func writeMessage(conn *websocket.Conn, receivedPackets <-chan HttpPacket) {
	for {
		receivedPacket := <-receivedPackets

		msg, err := json.Marshal(receivedPacket)
		if err != nil {
			return
		}

		err = conn.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			log.Printf("Websocket error: %s", err)
			conn.Close()
			return
		}
	}
}
