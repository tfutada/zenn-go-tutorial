package main

import (
	"fmt"
	"log"

	"github.com/google/gopacket/pcap"
)

// https://medium.com/a-bit-off/sniffing-network-go-6753cae91d3f
func main() {
	devices, err := pcap.FindAllDevs()
	if err != nil {
		log.Fatalf("error retrieving devices - %v", err)
	}

	for _, device := range devices {
		fmt.Printf("Device Name: %s\n", device.Name)
		fmt.Printf("Device Description: %s\n", device.Description)
		fmt.Printf("Device Flags: %d\n", device.Flags)
		for _, iaddress := range device.Addresses {
			fmt.Printf("\tInterface IP: %s\n", iaddress.IP)
			fmt.Printf("\tInterface NetMask: %s\n", iaddress.Netmask)
		}
	}
}
