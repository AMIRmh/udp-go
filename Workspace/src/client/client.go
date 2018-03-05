package client

import (
	"../../pkg/myLib"
	"fmt"
	"net"
	"strconv"
	"sync"
	"encoding/binary"
	"time"
)


const (
	PacketSize = 512
	PartSize = 3
)

var (
	waitAcksArray = make([]int, 0)
	threadNumber int
	udpAddr *net.UDPAddr
	conn *net.UDPConn
	ackArrayMutex sync.Mutex

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
	go getAck(data, part)
}

func getAck(data []byte, part int) {
	ackArrayMutex.Lock()
	waitAcksArray = append(waitAcksArray, part)
	ackArrayMutex.Unlock()
	time.Sleep(500 * time.Millisecond)

	if index := myLib.ContainsInt(waitAcksArray, part); index != myLib.Npos {
		ackArrayMutex.Lock()
		waitAcksArray = myLib.RemoveInt(waitAcksArray, index)
		ackArrayMutex.Unlock()
		sendUDP(data, part)
	}
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


