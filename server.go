package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/tcpassembly"
	"github.com/google/gopacket/tcpassembly/tcpreader"
	"github.com/gorilla/websocket"
)

func packetCapture(result chan<- *httpMessage) {
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

	streamFactory := &httpStreamFactory{result}
	streamPool := tcpassembly.NewStreamPool(streamFactory)
	assembler := tcpassembly.NewAssembler(streamPool)

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {
		if err := packet.ErrorLayer(); err != nil {
			log.Println("Error decoding some part of the packet:", err)
			continue
		}
		if packet.NetworkLayer() == nil || packet.TransportLayer() == nil ||
			packet.TransportLayer().LayerType() != layers.LayerTypeTCP {
			continue
		}

		flow := packet.NetworkLayer().NetworkFlow()
		tcp := packet.TransportLayer().(*layers.TCP)
		assembler.Assemble(flow, tcp)
	}
}

type httpMessage struct {
	AddressFlow string `json:"ip"`
	PortFlow    string `json:"port"`
	Raw         string `json:"raw"`
	Parsed      string `json:"payload"`
}

type httpStreamFactory struct {
	result chan<- *httpMessage
}

type httpStream struct {
	net, transport gopacket.Flow
	r              tcpreader.ReaderStream
	result         chan<- *httpMessage
}

func (h *httpStreamFactory) New(net, transport gopacket.Flow) tcpassembly.Stream {
	hstream := &httpStream{
		net:       net,
		transport: transport,
		r:         tcpreader.NewReaderStream(),
		result:    h.result,
	}
	go hstream.run()

	return &hstream.r
}

func (h *httpStream) run() {
	req, err := ioutil.ReadAll(&h.r)
	if err != nil {
		tcpreader.DiscardBytesToEOF(&h.r)
		log.Println("Error reading stream", h.net, h.transport, ":", err)
	}

	h.result <- &httpMessage{
		AddressFlow: fmt.Sprintf("%s", h.net),
		PortFlow:    fmt.Sprintf("%s", h.transport),
		Raw:         string(req),
		Parsed:      string(req),
	}
}

type wsClient struct {
	conn        *websocket.Conn
	coordinator *wsCoordinator
	send        chan *httpMessage
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
	broadcast   chan *httpMessage
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

		case httpMessage := <-coordinator.broadcast:
			for client := range coordinator.clients {
				client.send <- httpMessage
			}
		}
	}
}

func main() {
	coordinator := &wsCoordinator{
		clients:     make(map[*wsClient]bool),
		subscribe:   make(chan *wsClient),
		unsubscribe: make(chan *wsClient),
		broadcast:   make(chan *httpMessage),
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
			send:        make(chan *httpMessage),
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
