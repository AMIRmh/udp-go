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

	//conns = make([]*net.UDPConn, nth)
	//for i := 0; i < nth; i++ {
	//	newConn, err := net.DialUDP("udp", nil, udpAddr)
	//	myLib.CheckError(err)
	//	conns = append(conns, newConn)
	//}

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
	time.Sleep(500 * time.Millisecond)
	go connection.Read(id)
	if string(id) == DefaultId {
		go fillId()
	}
}


func Send(data []byte) {
	sendSize(len(data))
	sendChunk(data)
}


func sendChunk(input []byte) {
	var appendArray []byte
	if len(input) > DataSize {
		appendArray = make([]byte, len(input)%DataSize)
	} else {
		appendArray = make([]byte, DataSize-len(input))
	}
	input = append(input, appendArray...)
	parts := len(input) / DataSize
	fmt.Println("parts: ", len(input))

	var wg sync.WaitGroup
	wg.Add(numberOfThreads)
	for i := 0; i < numberOfThreads; i++ {
		go sendThreadParts(input, i, parts, &wg)
	}
	wg.Wait()

	// end message to close the connection.
	syncSendUDP([]byte("end"), specialMessage, createConnection())
	fmt.Println("finished sending")
}

func sendThreadParts(data []byte, threadId int, parts int, wg *sync.WaitGroup) {
	defer wg.Done()
	conn := createConnection()
	for i := threadId; i < parts; i+=numberOfThreads {
		//go addPartToWaitAckArray(data[i*DataSize: (i+1)*DataSize], i, conns[threadId])
		syncSendUDP(data[i*DataSize: (i+1)*DataSize], i, conn)
	}

	// get acks runs in a parallel loop with the program.
	finish := make(chan int, 1)
	go getAck(parts, finish, conn)

	// waits to finish the acks
	for range finish {}
	fmt.Println("thread " , threadId , " finished")
}

func asyncSendUDP(dataUdp []byte, part int, udpConn *net.UDPConn) {
	arr := make([]byte, PartSize)
	binary.BigEndian.PutUint32(arr, uint32(part))
	arr = append(arr, dataUdp...)
	data := id
	data = append(data, arr...)
	_, err := udpConn.Write(data)
	if part == specialMessage {
		fmt.Println(string(data))
		fmt.Println(data)
	}
	if err != nil {
		fmt.Println(err)
	}
}

func syncSendUDP(data []byte, part int, udpConn *net.UDPConn) {
	go addPartToWaitAckArray(data, part, udpConn)
	asyncSendUDP(data, part, udpConn)

	finish := make(chan int, 1)
	go getAck(1, finish, udpConn)
	for range finish {}
}

func addPartToWaitAckArray(data []byte, part int, udpConn *net.UDPConn) {
	ackArrayMutex.Lock()
	waitAcksArray = append(waitAcksArray, part)
	ackArrayMutex.Unlock()
	time.Sleep(500 * time.Millisecond)

	// maybe I should put lock here!!!!
	if removeElementFromAckArray(part) {
		asyncSendUDP(data, part, udpConn)
		go addPartToWaitAckArray(data, part, udpConn)
	}
}

func sendSize(size int) {
	fmt.Println("size: " + strconv.Itoa(size))
	syncSendUDP([]byte(strconv.Itoa(size)), specialMessage, createConnection())
	fmt.Println("send size finished")
}

func getAck(parts int, finish chan int, udpConn *net.UDPConn) {
	buf := make([]byte, 10)
	var mx sync.Mutex
	for i := 0; i < parts ; {
		udpConn.Read(buf[0:])
		go func() {
			part := binary.BigEndian.Uint32(buf)
			//myLib.CheckError(err)s
			// TODO I should find another way. mutex is not a good 	solution
			if removeElementFromAckArray(int(part)) {
				mx.Lock()
				i++
				mx.Unlock()
				if i >= parts {
					udpConn.Close()
				}
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


func createConnection() *net.UDPConn {
	connection, err := net.DialUDP("udp", nil, udpAddr)
	myLib.CheckError(err)
	return connection
}