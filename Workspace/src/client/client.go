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
	numberOfThreads int
	udpAddr *net.UDPAddr
	conns []*net.UDPConn
	ackArrayMutex sync.Mutex
	id  = make([]byte, 10)
)

func InitClient(host ,port string,nth int) {
	id = []byte(DefaultId)
	numberOfThreads = nth
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
	fmt.Println("sending chuncks")
	sendChunk(data)
}

func sendChunk(input []byte) {
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

	var wg sync.WaitGroup
	wg.Add(numberOfThreads)
	for i := 0; i < numberOfThreads; i++ {
		go sendThreadParts(input, i, parts, &wg)
	}
	wg.Wait()

	// end message to close the connection.
	fmt.Println("sending end to server")
	syncSendUDP([]byte("end"), specialMessage, createConnection())
	fmt.Println("finished sending")
}

func sendThreadParts(data []byte, threadId int, parts int, wgThread *sync.WaitGroup) {
	defer wgThread.Done()
	if threadId >= parts {
		return
	}
	var wg sync.WaitGroup
	for i := threadId; i < parts; i+=numberOfThreads {
		conn := createConnection()
		asyncSendUDP(data[i*DataSize: (i+1)*DataSize], i, conn)
		wg.Add(1)
		go retrySendUDPCommon(data[i*DataSize: (i+1)*DataSize], i, conn, &wg)
	}
	wg.Wait()
	fmt.Println("thread ", threadId, " finished")
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
	retrySendUDPCommon(data, part, udpConn, &wg)
	wg.Wait()
}

func retrySendUDPCommon(data []byte, part int, udpConn *net.UDPConn, wg *sync.WaitGroup) {
	buf := make([]byte, 10)
	udpConn.SetReadDeadline(time.Now().Add(ReadTimeOut))
	_, err := udpConn.Read(buf[0:])

	if err != nil {
		asyncSendUDP(data, part, udpConn)
		go retrySendUDPCommon(data, part, udpConn, wg)
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