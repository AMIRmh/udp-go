package main

import (
	"fmt"
	"net"
)

const host_server = "localhost"
const port_server = ":1313"

func main() {
	udpAddr , err := net.ResolveUDPAddr("udp4", port_server)

	if err != nil {
		fmt.Println(err)
		return
	}

	pc, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		fmt.Println(err)
		return
	}


	defer pc.Close()

	buffer := make([]byte, 1024)
	_, addr, err := pc.ReadFromUDP(buffer[0:])
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(string(buffer))

	pc.WriteToUDP([]byte("salam from client"), addr)
}