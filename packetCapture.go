package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

func packetCapture(result chan<- HttpPacket) {
	var (
		device       string        = "en0"
		snapshot_len int32         = 1024
		promiscuous  bool          = false
		timeout      time.Duration = 30 * time.Second
		filter       string        = "tcp or udp"
		handle       *pcap.Handle
		err          error
	)

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
			if strings.Contains(string(applicationLayer.Payload()), "HTTP") {

				var httpPacket HttpPacket

				ipLayer := packetData.Layer(layers.LayerTypeIPv4)
				if ipLayer != nil {
					ip, _ := ipLayer.(*layers.IPv4)
					httpPacket.source = fmt.Sprintf("%s", ip.SrcIP)
					httpPacket.destination = fmt.Sprintf("%s", ip.DstIP)
					httpPacket.protocol = fmt.Sprintf("%s", ip.Protocol)
				}

				tcpLayer := packetData.Layer(layers.LayerTypeTCP)
				if tcpLayer != nil {
					tcp, _ := tcpLayer.(*layers.TCP)
					httpPacket.source = fmt.Sprintf("%s:%s", httpPacket.source, tcp.SrcPort)
					httpPacket.destination = fmt.Sprintf("%s:%s", httpPacket.destination, tcp.DstPort)
				}

				httpPacket.payload = string(applicationLayer.Payload())

				result <- httpPacket
			}
		}

		if err := packetData.ErrorLayer(); err != nil {
			fmt.Println("Error decoding some part of the packet:", err)
		}
	}
}
