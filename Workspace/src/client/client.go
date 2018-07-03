package client

import (
	"fmt"
	"net"
	"strconv"
	"sync"
	"encoding/binary"
	"time"
	"udp-go/Workspace/pkg/myLib"
)

const (
	DataSize = 1500
	PartSize = 4
	DefaultId = "1234567890"
	ReadTimeOut = 500 * time.Millisecond
)

var (
	specialMessage = 4294967295
	waitAcksArray = make([]int, 0)
	windowsSize int
	udpAddr *net.UDPAddr
	conns []*net.UDPConn
	ackArrayMutex sync.Mutex
	id  = make([]byte, 10)
)

type Packet struct {
	partNumber int
	data []byte
}

func InitClient(host ,port string,nth int) {
	id = []byte(DefaultId)
	windowsSize = nth
	service := host + ":" + port
	udpAddr, _ = net.ResolveUDPAddr("udp4", service)
	getId()
}

func getId() {
	go fillId()
	for {
		if string(id) == DefaultId {
			fmt.Println("retying to connect")
			time.Sleep(1000 * time.Millisecond)
		} else {
			break
		}
	}
}

func fillId() {
	connection := createConnection()
	asyncSendUDP(id, specialMessage, connection)
	connection.SetReadDeadline(time.Now().Add(ReadTimeOut))
	connection.Read(id)
	if string(id) == DefaultId {
		fillId()
	}
}

func Send(data []byte) {
	fmt.Println("sending size")
	sendSize(len(data))
	fmt.Println("s=ending chuncks")
	sendChunk(data)
}

func sendChunk(input []byte) {
	// padding
	var appendArray []byte
	if len(input) > DataSize {
		appendArray = make([]byte, DataSize-len(input)%DataSize)
	} else {
		appendArray = make([]byte, DataSize-len(input))
	}
	fmt.Println(len(input) % DataSize)
	input = append(input, appendArray...)
	parts := len(input) / DataSize
	fmt.Println(len(input) % DataSize)
	
	// queue created
	packetsChannel := make(chan Packet, parts)
	go partsQueue(input, packetsChannel, parts)
	
	// sending parts in windows size
	var wg sync.WaitGroup
	wg.Add(windowsSize)
	for i := 0; i < windowsSize; i++ {
		go sendThreadParts(packetsChannel, &wg)
	}
	wg.Wait()

	// end message to close the connection.
	fmt.Println("sending end to server")
	syncSendUDP([]byte("end"), specialMessage, createConnection())
	fmt.Println("finished sending")
}

func partsQueue(input []byte, ch chan Packet, parts int) {
	for i := 0; i < parts; i++ {
		ch <- Packet{i, input[i * DataSize: (i + 1) * DataSize]}
	}
	close(ch)
}

func sendThreadParts(ch chan Packet, wgThread *sync.WaitGroup) {
	defer wgThread.Done()
	conn := createConnection()
	for packet := range ch {
		syncSendUDP(packet.data, packet.partNumber, conn)
	}
	fmt.Println("send part finished")
}

func asyncSendUDP(dataUdp []byte, part int, udpConn *net.UDPConn) {
	arr := make([]byte, PartSize)
	binary.BigEndian.PutUint32(arr, uint32(part))
	arr = append(arr, dataUdp...)
	data := id
	data = append(data, arr...)
	_, err := udpConn.Write(data)
	if err != nil {
		fmt.Println("here1")
		fmt.Println(err)
	}
}

func syncSendUDP(data []byte, part int, udpConn *net.UDPConn) {
	asyncSendUDP(data, part, udpConn)
	var wg sync.WaitGroup
	wg.Add(1)
	retrySendUDP(data, part, udpConn, &wg)
	wg.Wait()
}

func retrySendUDP(data []byte, part int, udpConn *net.UDPConn, wg *sync.WaitGroup) {
	buf := make([]byte, 10)
	udpConn.SetReadDeadline(time.Now().Add(ReadTimeOut))
	_, err := udpConn.Read(buf[0:])

	if err != nil {
		asyncSendUDP(data, part, udpConn)
		go retrySendUDP(data, part, udpConn, wg)
	} else {
		wg.Done()
	}
}

func sendSize(size int) {
	syncSendUDP([]byte(strconv.Itoa(size)), specialMessage, createConnection())
}

func createConnection() *net.UDPConn {
	connection, err := net.DialUDP("udp", nil, udpAddr)
	myLib.CheckError(err)
	return connection
}