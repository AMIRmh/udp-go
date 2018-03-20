package main

import (
	"fmt"
	"net"
	"strconv"
	"io/ioutil"
	"sync"
	"encoding/binary"
	"udp-go/Workspace/pkg/myLib"
)

const (
	hostServer = "localhost"
	DefaultId = "1234567890"
	PortServer = ":1313"
	DataSize = 1500
	IdSize = 10
	PartSize = 4
)

var (
	specialMessage = 4294967295
	udpAddr *net.UDPAddr
	pc *net.UDPConn
	arr = make([]string, 0)
	clientMutex = make(map[string]sync.Mutex)
	clientFiles =  make(map[string][]byte)
	clientAckAddrs = make(map[string]*net.UDPAddr)
)

func main() {
	initServer()
	readUDP()
}

func initServer() {
	udpAddr , err := net.ResolveUDPAddr("udp4", PortServer)

	if err != nil {
		fmt.Println(err)
		return
	}

	pc, err = net.ListenUDP("udp", udpAddr)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func readUDP() {
	buf := make([]byte, DataSize+PartSize+IdSize)
	for {
		_, remoteAddr, _ := pc.ReadFromUDP(buf[0:])
		go processUDP(buf, remoteAddr)
		buf = make([]byte, DataSize+PartSize+IdSize)
	}
}

func processUDP(buffer []byte, remoteAddr *net.UDPAddr) {
	id := buffer[0:IdSize]
	partBuffer := buffer[IdSize:IdSize+PartSize]
	part := int(binary.BigEndian.Uint32(partBuffer))
	data := buffer[IdSize+PartSize:]

	if part == specialMessage {
		specialMessageHandler(data, id, partBuffer, remoteAddr)
	} else {
		putInArray(id, data, part)
		pc.WriteToUDP(partBuffer, clientAckAddrs[string(id)])
	}
}

func specialMessageHandler(data, id, partBuffer []byte, remoteAddr *net.UDPAddr) {

	fmt.Println(string(data))
	if trimNullString(data) == "introduceAck" {

		clientAckAddrs[string(id)] = remoteAddr
		pc.WriteToUDP(partBuffer, remoteAddr)

	} else if trimNullString(data) == "end" {

		fmt.Println("end recieved")
		ioutil.WriteFile("./a", clientFiles[string(id)], 0777)
		delete(clientFiles, string(id))
		pc.WriteToUDP(partBuffer, remoteAddr)

	} else if string(id) == DefaultId {

		newId := myLib.RandStringRunes(IdSize)
		var mx sync.Mutex
		clientMutex[string(newId)] = mx
		pc.WriteToUDP([]byte(newId), remoteAddr)

	} else if size, err := strconv.Atoi(trimNullString(data)); err == nil {

		if size < DataSize {
			clientFiles[string(id)] = make([]byte, DataSize)
		} else {
			clientFiles[string(id)] = make([]byte, size + size%DataSize)
		}
		pc.WriteToUDP(partBuffer, remoteAddr)

	}
}

func putInArray(id, data []byte, part int) {
	fmt.Println(part)

	mx := clientMutex[string(id)]
	clientPart := clientFiles[string(id)]
	clientPart = append(clientPart[0:part*DataSize],
					append(data, clientPart[(part+1)*DataSize:]...)...)

	mx.Lock()
	clientFiles[string(id)] = clientPart
	mx.Unlock()

	fmt.Println(clientFiles[string(id)])

}

func trimNullString(str []byte) string {
	var index int
	for i, s := range str {
		if s == byte(0) {
			index = i
			break
		}
	}
	return string(str[0:index])
}