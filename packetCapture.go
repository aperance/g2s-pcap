package main

import (
	"encoding/json"
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
	filter       string        = "tcp port 80 or udp"
	handle       *pcap.Handle
	err          error
)

type HttpPacket struct {
	Source      string `json:"src"`
	Destination string `json:"dst"`
	Protocol    string `json:"type"`
	Payload     string `json:"payload"`
	Push        bool   `json:"push"`
}

func packetCapture(result chan<- []byte) {
	log.Println("Capturing HTTP packets...")

	handle, err = pcap.OpenLive(device, snapshot_len, promiscuous, timeout)
	if err != nil {
		log.Fatal(err)
	}
	defer handle.Close()

	err = handle.SetBPFFilter(filter)
	if err != nil {
		log.Fatal(err)
	}

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())

	for packetData := range packetSource.Packets() {

		applicationLayer := packetData.ApplicationLayer()
		if applicationLayer != nil {

			var httpPacket HttpPacket

			ipLayer := packetData.Layer(layers.LayerTypeIPv4)
			if ipLayer != nil {
				ip, _ := ipLayer.(*layers.IPv4)

				tcpLayer := packetData.Layer(layers.LayerTypeTCP)
				if tcpLayer != nil {
					tcp, _ := tcpLayer.(*layers.TCP)
					httpPacket.Source = fmt.Sprintf("%s:%s", ip.SrcIP, tcp.SrcPort)
					httpPacket.Destination = fmt.Sprintf("%s:%s", ip.DstIP, tcp.DstPort)
					httpPacket.Push = tcp.PSH
				}

				udpLayer := packetData.Layer(layers.LayerTypeUDP)
				if udpLayer != nil {
					udp, _ := udpLayer.(*layers.UDP)
					httpPacket.Source = fmt.Sprintf("%s:%s", ip.SrcIP, udp.SrcPort)
					httpPacket.Destination = fmt.Sprintf("%s:%s", ip.DstIP, udp.DstPort)
					httpPacket.Push = true
				}

				httpPacket.Protocol = fmt.Sprintf("%s", ip.Protocol)
				httpPacket.Payload = string(applicationLayer.Payload())

				msg, err := json.Marshal(httpPacket)
				if err != nil {
					return
				}

				result <- msg

			}

		}

		if err := packetData.ErrorLayer(); err != nil {
			fmt.Println("Error decoding some part of the packet:", err)
		}
	}
}
