package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/gorilla/websocket"
)

type message struct {
	Flow      string `json:"flow"`
	Protocol  string `json:"protocol"`
	Direction string `json:"direction"`
	Payload   string `json:"payload"`
}

func packetCapture(result chan<- *message) {
	var port, device string
	flag.StringVar(&port, "p", "47028", "port number")
	flag.StringVar(&device, "d", "\\Device\\NPF_{B82CD492-3820-4FCD-9309-552CA055A24C}", "interface device name")
	flag.Parse()

	handle, err := pcap.OpenLive(device, 2048, false, 30*time.Second)
	if err != nil {
		log.Fatal(err)
	}
	defer handle.Close()

	err = handle.SetBPFFilter("udp or tcp port " + port)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Capturing HTTP packets on port %s...", port)

	startTagRegex := regexp.MustCompile(`<[Ss][^<]*?:Body.*?>\s*`)
	endTagRegex := regexp.MustCompile(`\s*<\/[Ss][^<]*?:Body.*?>`)

	buffer := make(map[string]([]byte))
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())

	for packet := range packetSource.Packets() {
		if err := packet.ErrorLayer(); err != nil {
			log.Println("Error decoding some part of the packet:", err)
			continue
		}
		if packet.NetworkLayer() == nil || packet.TransportLayer() == nil {
			continue
		}

		networkFlow := packet.NetworkLayer().NetworkFlow()
		transportFlow := packet.TransportLayer().TransportFlow()
		flowString := networkFlow.Src().String() +
			":" + transportFlow.Src().String() +
			" -> " + networkFlow.Dst().String() +
			":" + transportFlow.Dst().String()

		if packet.TransportLayer().LayerType() == layers.LayerTypeUDP {
			payload := packet.TransportLayer().(*layers.UDP).Payload

			if fmt.Sprintf("%x", payload[0:1]) == "a4" {
				direction := ""
				if networkFlow.Dst().String() == "172.20.109.41" {
					direction = "Outbound"
				} else {
					direction = "Inbound"
				}

				result <- &message{
					Flow:      flowString,
					Protocol:  "Freeform",
					Direction: direction,
					Payload:   fmt.Sprintf("%x", payload),
				}
			}
		}

		if packet.TransportLayer().LayerType() == layers.LayerTypeTCP {
			if packet.ApplicationLayer() == nil {
				continue
			}

			key := flowString
			buffer[key] = append(buffer[key], packet.ApplicationLayer().Payload()...)

			if packet.TransportLayer().(*layers.TCP).PSH {
				startTagLocation := startTagRegex.FindIndex(buffer[key])
				endTagLocation := endTagRegex.FindIndex(buffer[key])

				if startTagLocation != nil && endTagLocation != nil {
					direction := ""
					if networkFlow.Src().String() == "172.20.109.46" {
						direction = "Outbound"
					} else {
						direction = "Inbound"
					}

					payload := buffer[key][startTagLocation[1]:endTagLocation[0]]

					result <- &message{
						Flow:      flowString,
						Protocol:  "G2S",
						Direction: direction,
						Payload:   string(payload),
					}
				}

				delete(buffer, key)
			}

		}
	}
}

type wsClient struct {
	conn        *websocket.Conn
	coordinator *wsCoordinator
	send        chan *message
}

func (client *wsClient) read() {
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

func (client *wsClient) write() {
	for {
		err := client.conn.WriteJSON(<-client.send)
		if err != nil {
			log.Println(err)
			return
		}
	}
}

type wsCoordinator struct {
	clients     map[*wsClient]bool
	subscribe   chan *wsClient
	unsubscribe chan *wsClient
	broadcast   chan *message
}

func (coordinator *wsCoordinator) run() {
	for {
		select {
		case client := <-coordinator.subscribe:
			coordinator.clients[client] = true

		case client := <-coordinator.unsubscribe:
			if _, ok := coordinator.clients[client]; ok {
				delete(coordinator.clients, client)
			}

		case message := <-coordinator.broadcast:
			for client := range coordinator.clients {
				client.send <- message
			}
		}
	}
}

func main() {
	coordinator := &wsCoordinator{
		clients:     make(map[*wsClient]bool),
		subscribe:   make(chan *wsClient),
		unsubscribe: make(chan *wsClient),
		broadcast:   make(chan *message),
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

		client := &wsClient{
			conn:        conn,
			coordinator: coordinator,
			send:        make(chan *message),
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
