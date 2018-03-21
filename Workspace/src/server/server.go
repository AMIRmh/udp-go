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
	clienSizes = make(map[string]int)
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
		defer func() {
			if r := recover(); r == nil {
				fmt.Println("sending part")
				pc.WriteToUDP(partBuffer, clientAckAddrs[string(id)])
			} else {
				fmt.Println("panic happend but recovered!!!!")
			}
		}()
		fmt.Println("part came: ", part)
		putInArray(id, data, part)
	}
}

func specialMessageHandler(data, id, partBuffer []byte, remoteAddr *net.UDPAddr) {

	if trimNullString(data) == "introduceAck" {

		clientAckAddrs[string(id)] = remoteAddr
		pc.WriteToUDP(partBuffer, remoteAddr)

	} else if trimNullString(data) == "end" {

		fmt.Println(string(id), " end recieved")
		ioutil.WriteFile("./a", clientFiles[string(id)][0:clienSizes[string(id)]], 0777)
		delete(clientFiles, string(id))
		delete(clienSizes, string(id))
		delete(clientAckAddrs, string(id))
		pc.WriteToUDP(partBuffer, remoteAddr)

	} else if string(id) == DefaultId {

		newId := myLib.RandStringRunes(IdSize)
		var mx sync.Mutex
		clientMutex[string(newId)] = mx
		pc.WriteToUDP([]byte(newId), remoteAddr)

	} else if size, err := strconv.Atoi(trimNullString(data)); err == nil {

		clienSizes[string(id)] = size


		if size < DataSize {
			clientFiles[string(id)] = make([]byte, DataSize)
		} else {
			clientFiles[string(id)] = make([]byte, size + (DataSize-size%DataSize))
			fmt.Println(len(clientFiles[string(id)]))
		}
		pc.WriteToUDP(partBuffer, remoteAddr)
	}
}

func putInArray(id, data []byte, part int) {
	mx := clientMutex[string(id)]
	clientPart := clientFiles[string(id)]
	clientPart = append(clientPart[0:part*DataSize],
					append(data, clientPart[(part+1)*DataSize:]...)...)

	mx.Lock()
	clientFiles[string(id)] = clientPart
	mx.Unlock()

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