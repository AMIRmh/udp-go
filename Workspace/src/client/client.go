package client

import (
	"../../pkg/myLib"
	"fmt"
	"net"
	"strconv"
	"sync"
	"encoding/binary"
)


const (
	PacketSize = 512
	PartSize = 3
)

var (
	 threadNumber int
	 udpAddr *net.UDPAddr
	 conn *net.UDPConn
	 udpMutex sync.Mutex
)

func InitClient(host ,port string,thn int) {
	threadNumber = thn
	service := host + ":" + port

	udpAddr, err := net.ResolveUDPAddr("udp4", service)
	myLib.CheckError(err)

	conn, err = net.DialUDP("udp", nil, udpAddr)
	myLib.CheckError(err)
}


func Send(data []byte) {
	sendChunk(data)
}


func sendChunk(data []byte) {
	fmt.Println(len(data))
	//parts := len(data) / PACKET_SIZE
	for i := 0; i < 500; i++ {
		conn.Write([]byte(strconv.Itoa(i)))
	}

	sendSize(len(data))
}

func sendParts() {

}

func sendThreadParts() {

}

func sendUDP(data []byte, part int) {
	arr := make([]byte, PartSize)
	binary.LittleEndian.PutUint16(arr, uint16(part))
	arr = myLib.Reverse(arr)
	arr = append(arr, data...)
	conn.Write(arr)

}

func getAck() {

}

func sendSize(size int) {
	fmt.Println(size)
	_, err := conn.Write([]byte(strconv.Itoa(size)))
	myLib.CheckError(err)
	var buf [PacketSize]byte
	n, err := conn.Read(buf[0:])
	myLib.CheckError(err)
	fmt.Println(string(buf[0:n]))
}


