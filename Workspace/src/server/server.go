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
	portServer = ":1313"
	DataSize = 1500
	idSize = 10
	partSize = 4
)

var (
	specialMessage = 4294967295
	udpAddr *net.UDPAddr
	pc *net.UDPConn
	arr = make([]string, 0)
	clientMutex = make(map[string]sync.Mutex)
	clientFiles =  make(map[string][]byte)
)

func main() {
	initServer()
	readUDP()
}

func initServer() {
	udpAddr , err := net.ResolveUDPAddr("udp4", portServer)

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
	buf := make([]byte, DataSize+partSize+idSize)
	fmt.Println(byte(0))
	for {
		_, remoteAddr, _ := pc.ReadFromUDP(buf[0:])
		go processUDP(buf, remoteAddr)
		buf = make([]byte, DataSize+partSize+idSize)
	}
}

func processUDP(buffer []byte, remoteAddr *net.UDPAddr) {
	id := buffer[0:idSize]
	partBuffer := buffer[idSize:idSize+partSize]
	part := binary.BigEndian.Uint32(partBuffer)
	data := buffer[idSize+partSize:]

	if part == uint32(specialMessage) {
		if string(data) == "end" {
			ioutil.WriteFile("./a", clientFiles[string(id)], 0777)
			delete(clientFiles, string(id))
			pc.WriteToUDP(partBuffer, remoteAddr)
		} else if string(id) == DefaultId {
			newId := myLib.RandStringRunes(idSize)
			var mx sync.Mutex
			clientMutex[string(newId)] = mx
			pc.WriteToUDP([]byte(newId), remoteAddr)
		} else if n, err := strconv.Atoi(trimNullString(data)); err == nil {

			clientFiles[string(id)] = make([]byte, n)
			pc.WriteToUDP(partBuffer, remoteAddr)
		}

		return
	}

	putInArray(id, data, int(part))
	pc.Write(partBuffer)
}

func putInArray(id, data []byte, part int) {
	mx := clientMutex[string(id)]
	mx.Lock()
	fmt.Println("this is :" , len(clientFiles[string(id)]))
	clientPart := clientFiles[string(id)]
	clientPart = append(clientPart[0:part*DataSize],
					append(data, clientPart[(part+1)*DataSize:]...)...)
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