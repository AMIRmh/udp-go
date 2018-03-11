package server

import (
	"fmt"
	"net"
	"strconv"
	"io/ioutil"
	"sync"
	"os"
	"encoding/binary"
	"reflect"
)

const (
	hostServer = "localhost"
	portServer = ":1313"
	PacketSize = 1500
	idSize = 10
	partSize = 3
)

var (
	endMessage = 4294967295
	udpAddr *net.UDPAddr
	pc *net.UDPConn
	arr = make([]string, 0)
	m sync.Mutex
	clientFiles map[string][]byte
)

func main() {
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


	defer pc.Close()

	var wg sync.WaitGroup

	wg.Add(500)
	for i := 0; i < 500; i++ {
		go poc(i, &wg)
	}
	wg.Wait()
	fmt.Println(arr)
	fmt.Println(len(arr))
	os.Exit(0)

	buffer := make([]byte, PacketSize)
	_, addr, err := pc.ReadFromUDP(buffer[0:])
	size, _ := strconv.Atoi(string(buffer))
	file := make([]byte, size)

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("size: " + string(buffer))
	//size,_ := strconv.Atoi(string(buffer))
	//arr := make([]byte, size)


	pc.WriteToUDP([]byte("salam from client"), addr)
	pc.ReadFromUDP(file)
	ioutil.WriteFile("./a", file, 0777)
}


func poc(i int, wg *sync.WaitGroup) {
	defer wg.Done()
	buffer := make([]byte, 1024)
	pc.Read(buffer)
	fmt.Println("got " + strconv.Itoa(i) + " " + string(buffer))
	m.Lock()
	arr = append(arr, string(buffer))
	m.Unlock()
}


func readUDP() {
	buf := make([]byte, 2000)

	for {
		pc.ReadFromUDP(buf[0:])
		copyPc := pc
		go processUDP(buf, copyPc)
	}

}

func processUDP(buffer []byte, pc *net.UDPConn) {

	id := buffer[:idSize]
	partBuffer := buffer[idSize:idSize+partSize]
	part := binary.BigEndian.Uint32(partBuffer)
	data := buffer[idSize+partSize:]

	if part == uint32(endMessage) {
		ioutil.WriteFile("./a", clientFiles[string(id)], 0777)
		delete(clientFiles, string(id))
		pc.Write(partBuffer)
		return
	}


	putInArray(id, data, int(part))
	pc.Write(partBuffer)
}

func putInArray(id, data []byte, part int) {
	clientFiles[string(id)][part*PacketSize:(part+1)*PacketSize] = data
}