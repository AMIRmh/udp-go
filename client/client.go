package client

import (
	"fmt"
	"net"
	"os"
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
	checkError(err)

	conn, err = net.DialUDP("udp", nil, udpAddr)
	checkError(err)
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

func sendUDP(data []byte, part uint16) {
	arr := make([]byte, PartSize)
	binary.LittleEndian.PutUint16(arr, part)
	arr = reverse(arr)
	arr = append(arr, data...)

}


func reverse(s []byte) []byte {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}

func getAck() {

}

func sendSize(size int) {
	fmt.Println(size)
	_, err := conn.Write([]byte(strconv.Itoa(size)))
	checkError(err)
	var buf [PacketSize]byte
	n, err := conn.Read(buf[0:])
	checkError(err)
	fmt.Println(string(buf[0:n]))
}


func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error ", err.Error())
		os.Exit(1)
	}
}