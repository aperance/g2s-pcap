package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/gorilla/websocket"
)

var (
	port   string
	device string
)

type HttpPacket struct {
	Source      string `json:"src"`
	Destination string `json:"dst"`
	Payload     string `json:"payload"`
}

func packetCapture(result chan<- *HttpPacket) {
	var port, device string
	flag.StringVar(&port, "p", "", "port number")
	flag.StringVar(&device, "d", "", "interface device name")
	flag.Parse()

	handle, err := pcap.OpenLive(device, 2048, false, 30*time.Second)
	if err != nil {
		log.Fatal(err)
	}
	defer handle.Close()

	err = handle.SetBPFFilter("tcp port " + port)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Capturing HTTP packets on port %s...", port)

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	buffer := make(map[string]string)

	for packetData := range packetSource.Packets() {
		if err := packetData.ErrorLayer(); err != nil {
			log.Println("Error decoding some part of the packet:", err)
			continue
		}

		ip, ok := packetData.Layer(layers.LayerTypeIPv4).(*layers.IPv4)
		if !ok || ip == nil {
			continue
		}

		tcp, ok := packetData.Layer(layers.LayerTypeTCP).(*layers.TCP)
		if !ok || tcp == nil {
			continue
		}

		applicationLayer := packetData.ApplicationLayer()
		if applicationLayer == nil {
			continue
		}

		key := fmt.Sprintf("%s:%s|%s:%s", ip.SrcIP, tcp.SrcPort, ip.DstIP, tcp.DstPort)
		buffer[key] = buffer[key] + string(applicationLayer.Payload())

		if tcp.PSH {
			httpPacket := HttpPacket{
				Source:      fmt.Sprintf("%s:%s", ip.SrcIP, tcp.SrcPort),
				Destination: fmt.Sprintf("%s:%s", ip.DstIP, tcp.DstPort),
				Payload:     buffer[key],
			}
			result <- &httpPacket

			delete(buffer, key)
		}
	}
}

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
		err := client.conn.WriteJSON(<-client.send)
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
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}

		log.Println("Websocket client connected")

		client := &Client{
			conn:        conn,
			coordinator: coordinator,
			send:        make(chan *HttpPacket),
		}

		go client.read()
		go client.write()

		client.coordinator.subscribe <- client
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "old_index.html")
	})

	http.ListenAndServe(":3000", nil)
}
