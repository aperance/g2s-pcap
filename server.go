package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/tcpassembly"
	"github.com/google/gopacket/tcpassembly/tcpreader"
	"github.com/gorilla/websocket"
)

type Message struct {
	AddressFlow string `json:"ip"`
	PortFlow    string `json:"port"`
	Protocol	string `json:"protocol"`
	Payload     string `json:"payload"`
}

func packetCapture( result chan<- *Message) {
	var port, device string
	flag.StringVar(&port, "p", "", "port number")
	flag.StringVar(&device, "d", "", "interface device name")
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

	streamFactory := &g2sStreamFactory{result}
	streamPool := tcpassembly.NewStreamPool(streamFactory)
	assembler := tcpassembly.NewAssembler(streamPool)

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {
		if err := packet.ErrorLayer(); err != nil {
			log.Println("Error decoding some part of the packet:", err)
			continue
		}
		if packet.NetworkLayer() == nil || packet.TransportLayer() == nil  {
			continue
		}
		if packet.TransportLayer().LayerType() == layers.LayerTypeUDP {
			udp := packet.TransportLayer().(*layers.UDP)
			if udp.SrcPort.String() != "49152" && udp.DstPort.String() != "49152" {
				continue
			}
			result <- &Message{
				AddressFlow:	packet.NetworkLayer().NetworkFlow().String(),
				PortFlow:		packet.TransportLayer().TransportFlow().String(),
				Protocol:		"freeform",
				Payload:		fmt.Sprintf("%x", udp.Payload),
			}
		}
		if packet.TransportLayer().LayerType() == layers.LayerTypeTCP {
			flow := packet.NetworkLayer().NetworkFlow()
			tcp := packet.TransportLayer().(*layers.TCP)
			assembler.Assemble(flow, tcp)
		}
	}
}

type g2sStreamFactory struct {
	result chan<- *Message
}

type g2sStream struct {
	net, transport gopacket.Flow
	r              tcpreader.ReaderStream
	result         chan<- *Message
}

func (f *g2sStreamFactory) New(net, transport gopacket.Flow) tcpassembly.Stream {
	stream := &g2sStream{
		net:       net,
		transport: transport,
		r:         tcpreader.NewReaderStream(),
		result:    f.result,
	}
	go stream.scan()

	return &stream.r
}

func (stream *g2sStream) scan() {
	openingTagRegex := regexp.MustCompile(`<s[^<]*?:Body.*?>\s*`)
	closingTagRegex := regexp.MustCompile(`\s*<\/s[^<]*?:Body.*?>`)

	scanner := bufio.NewScanner(&stream.r)
	split := func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF {
			return 0, nil, io.EOF
		}
		openingTag := openingTagRegex.FindIndex(data)
		if openingTag == nil {
			return 0, nil, nil
		}
		closingTag := closingTagRegex.FindIndex(data)
		if closingTag == nil {
			return 0, nil, nil
		}
		return closingTag[1], data[openingTag[1]:closingTag[0]], nil
	}
	scanner.Split(split)

	for scanner.Scan() {
		r := strings.NewReplacer(`&lt;?xml version="1.0"?&gt;`, "", "&lt;", "<", "&gt;", ">")
		stream.result <- &Message{
			AddressFlow:	fmt.Sprintf("%s", stream.net),
			PortFlow:		fmt.Sprintf("%s", stream.transport),
			Protocol:		"g2s",
			Payload:		r.Replace(scanner.Text()),
		}
	}
	if scanner.Err() != nil {
		fmt.Printf("error: %s\n", scanner.Err())
	}
}

type wsClient struct {
	conn        *websocket.Conn
	coordinator *wsCoordinator
	send        chan *Message
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
	broadcast   chan *Message
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

		case Message := <-coordinator.broadcast:
			for client := range coordinator.clients {
				client.send <- Message
			}
		}
	}
}

func main() {
	coordinator := &wsCoordinator{
		clients:     make(map[*wsClient]bool),
		subscribe:   make(chan *wsClient),
		unsubscribe: make(chan *wsClient),
		broadcast:   make(chan *Message),
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
			send:        make(chan *Message),
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
