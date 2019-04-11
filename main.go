package main

import (
	"fmt"
	"net/http"
)

type HttpPacket struct {
	source      string
	destination string
	protocol    string
	payload     string
}

func main() {
	receivedPackets := make(chan HttpPacket)

	go packetCapture(receivedPackets)
	fmt.Println("Capturing HTTP packets...")
	go webSocketServer()
	fmt.Println("Listening on port 3000...")
	for {
		receivedPacket := <-receivedPackets
		fmt.Printf("%s packet from %s to %s\n", receivedPacket.protocol, receivedPacket.source, receivedPacket.destination)
		fmt.Printf("\n%s\n\n", receivedPacket.payload)
	}
}

func webSocketServer() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello World!")
	})
	http.ListenAndServe(":3000", nil)
}
