package main

import (
	"fmt"
	"net"
	"strconv"
	"io/ioutil"
	"sync"
	"os"
)

const host_server = "localhost"
const port_server = ":1313"
const PACKET_SIZE = 512
var udpAddr *net.UDPAddr
var pc *net.UDPConn
var arr = make([]string, 0)
var m sync.Mutex
func main() {
	udpAddr , err := net.ResolveUDPAddr("udp4", port_server)

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

	buffer := make([]byte, PACKET_SIZE)
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