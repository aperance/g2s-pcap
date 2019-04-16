package main

import (
	"fmt"
	"log"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

var (
	device       string        = "en0"
	snapshot_len int32         = 1024
	promiscuous  bool          = false
	timeout      time.Duration = 30 * time.Second
	filter       string        = "tcp port"
	handle       *pcap.Handle
	err          error
)

type HttpPacket struct {
	Source      string `json:"src"`
	Destination string `json:"dst"`
	Payload     string `json:"payload"`
}

func packetCapture(port string, result chan<- *HttpPacket) {

	buffer := make(map[string]string)

	handle, err = pcap.OpenLive(device, snapshot_len, promiscuous, timeout)
	if err != nil {
		log.Fatal(err)
	}
	defer handle.Close()

	err = handle.SetBPFFilter("tcp port " + port)
	if err != nil {
		log.Fatal(err)
	}

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())

	log.Printf("Capturing HTTP packets on port %s...", port)

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
		buffer[key] = buffer[key] + "_@_" + string(applicationLayer.Payload())

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
