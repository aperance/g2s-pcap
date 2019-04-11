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

func packetCapture(result chan<- string) {
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

	for packet := range packetSource.Packets() {

		applicationLayer := packet.ApplicationLayer()
		if applicationLayer != nil {
			if strings.Contains(string(applicationLayer.Payload()), "HTTP") {

				ipLayer := packet.Layer(layers.LayerTypeIPv4)
				if ipLayer != nil {
					ip, _ := ipLayer.(*layers.IPv4)
					fmt.Printf("%s packet from %s to %s\n", ip.Protocol, ip.SrcIP, ip.DstIP)
				}

				tcpLayer := packet.Layer(layers.LayerTypeTCP)
				if tcpLayer != nil {
					tcp, _ := tcpLayer.(*layers.TCP)
					fmt.Printf("Port %d to %d\n", tcp.SrcPort, tcp.DstPort)
				}

				result <- string(applicationLayer.Payload())
			}
		}

		if err := packet.ErrorLayer(); err != nil {
			fmt.Println("Error decoding some part of the packet:", err)
		}
	}
}
