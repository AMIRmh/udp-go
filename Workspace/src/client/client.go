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
	PacketSize = 1500
	PartSize = 4
)

var (
	endMessage = 4294967295
	waitAcksArray = make([]int, 0)
	numberOfThreads int
	udpAddr *net.UDPAddr
	conn *net.UDPConn
	ackArrayMutex sync.Mutex
	id  = make([]byte, 10)
)

func InitClient(host ,port string,nth int) {
	id = []byte("null")
	numberOfThreads = nth
	service := host + ":" + port

	udpAddr, err := net.ResolveUDPAddr("udp4", service)
	myLib.CheckError(err)

	conn, err = net.DialUDP("udp", nil, udpAddr)
	myLib.CheckError(err)

	go conn.Read(id)
}

func getId() {
	conn.Write([]byte("id"))
	go getIdReader()
}

func getIdReader() {
	time.Sleep(500 * time.Millisecond)
	if string(id) == "null" {
		go getId()
	}
}


func Send(data []byte) {
	sendSize(len(data))
	sendChunk(data)
}


func sendChunk(input []byte) {
	data := make([]byte, len(input) + len(input) % PacketSize)
	data[:len(input)] = input

	parts := len(data) / PacketSize

	// get acks runs in a parallel loop with the program.
	finish := make(chan int, 1)
	go getAck(parts, finish)

	for i := 0; i < numberOfThreads; i++ {
		go sendThreadParts(data, i, parts)
	}

	// waits to finish the acks
	for range finish {}

	finish = make(chan int, 1)
	go getAck(1, finish)
	sendUDP([]byte(""), endMessage)

	fmt.Println("finished sending")
}

func sendThreadParts(data []byte, threadId int, parts int) {
	for i := threadId; i < parts; i+=numberOfThreads {
		sendUDP(data[i*PacketSize: (i+1)*PacketSize], i)
	}
}

func sendUDP(dataUdp []byte, part int) {
	arr := make([]byte, PartSize)
	binary.BigEndian.PutUint32(arr, uint32(part))
	arr = append(arr, dataUdp...)
	conn.Write(arr)
	go addPartToWaitAckArray(dataUdp, part)
}

func addPartToWaitAckArray(data []byte, part int) {
	ackArrayMutex.Lock()
	waitAcksArray = append(waitAcksArray, part)
	ackArrayMutex.Unlock()
	time.Sleep(500 * time.Millisecond)

	// maybe I should put lock here!!!!
	if removeElementFromAckArray(part) {
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

func getAck(parts int, finish chan int) {
	buf := make([]byte, 10)
	var mx sync.Mutex
	for i := 0; i < parts; {
		conn.Read(buf[0:])
		go func() {
			part, err := strconv.Atoi(string(buf[0:]))
			myLib.CheckError(err)
			// TODO I should find another way. mutex is not a good 	solution
			if removeElementFromAckArray(part) {
				mx.Lock()
				i++
				mx.Unlock()
			}
		}()
	}
	close(finish)
}

func removeElementFromAckArray(part int) bool {
	if index := myLib.ContainsInt(waitAcksArray, part); index != myLib.Npos {
		ackArrayMutex.Lock()
		waitAcksArray = myLib.RemoveInt(waitAcksArray, index)
		ackArrayMutex.Unlock()
		return true
	}
	return false
}
