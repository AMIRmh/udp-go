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
	fmt.Println("sending intro")
	conn := createConnection()
	introduceAckConnection(conn)
	//conn.SetReadDeadline(time.Now().Add(1000000 * time.Second))
	//_, err := conn.Read(make([]byte,10))
	//fmt.Println("err", err)
	fmt.Println("sending chuncks")
	sendChunk(data, conn)
}

func introduceAckConnection(conn *net.UDPConn ) {

	asyncSendUDP([]byte("introduceAck"), specialMessage, conn)
	buf := make([]byte, PartSize)
	conn.SetReadDeadline(time.Now().Add(ReadTimeOut))
	conn.Read(buf)
	if int(binary.BigEndian.Uint32(buf)) != specialMessage {
		introduceAckConnection(conn)
	}
}

func sendChunk(input []byte, conn *net.UDPConn) {
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

	for i := 0; i < numberOfThreads; i++ {
		go sendThreadParts(input, i, parts)
	}

	// get acks runs in a parallel loop with the program.
	finish := make(chan int, 1)
	go getAckCommon(parts, finish, conn)

	// waits to finish the acks
	for range finish {}

	// end message to close the connection.
	fmt.Println("sending end to server")
	syncSendUDP([]byte("end"), specialMessage, createConnection())
	fmt.Println("finished sending")
}

func sendThreadParts(data []byte, threadId int, parts int) {
	conn := createConnection()
	if threadId >= parts {
		return
	}
	var wg sync.WaitGroup
	for i := threadId; i < parts; i+=numberOfThreads {
		asyncSendUDP(data[i*DataSize: (i+1)*DataSize], i, conn)
		wg.Add(1)
		go retrySendUDPCommon(data[i*DataSize: (i+1)*DataSize], i, conn, &wg)
	}
	wg.Wait()
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
	ackArrayMutex.Lock()
	waitAcksArray = append(waitAcksArray, part)
	ackArrayMutex.Unlock()

	asyncSendUDP(data, part, udpConn)

	finish := make(chan int, 1)
	go getAckSpecial(data, part, 1, finish, udpConn)
	for range finish {}
}

func retrySendUDPCommon(data []byte, part int, udpConn *net.UDPConn, wg *sync.WaitGroup) {
	ackArrayMutex.Lock()
	waitAcksArray = append(waitAcksArray, part)
	ackArrayMutex.Unlock()
	time.Sleep(500 * time.Millisecond)

	if removeElementFromAckArray(part) {
		asyncSendUDP(data, part, udpConn)
		//go retrySendUDPCommon(data, part, udpConn)
		retrySendUDPCommon(data, part, udpConn, wg)
	} else {
		wg.Done()
	}
}

func sendSize(size int) {
	syncSendUDP([]byte(strconv.Itoa(size)), specialMessage, createConnection())
}

func getAckSpecial(data []byte, part int,  parts int, finish chan int, udpConn *net.UDPConn) {
	buf := make([]byte, 10)
	for i := 0; i < parts ; {
		udpConn.SetReadDeadline(time.Now().Add(ReadTimeOut))
		_, err := udpConn.Read(buf[0:])
		if err != nil {
			asyncSendUDP(data, part, udpConn)
		} else {
			part := binary.BigEndian.Uint32(buf)
			//fmt.Println("hey", part)
			//myLib.CheckError(err)s
			// TODO I should find another way. mutex is not a good 	solution
			if removeElementFromAckArray(int(part)) {
				i++
				if i >= parts {
					udpConn.Close()
				}
			}
		}
	}
	close(finish)
}


func getAckCommon(parts int, finish chan int, udpConn *net.UDPConn) {
	buf := make([]byte, 10)
	var mx sync.Mutex
	ii := 0
	for i := 0; i < parts ; {
		udpConn.SetReadDeadline(time.Now().Add(10000000 * time.Second))
		_, err := udpConn.Read(buf[0:])
		if err == nil {
			//fmt.Println(buf)
			go func(i *int) {
				part := binary.BigEndian.Uint32(buf)
				//myLib.CheckError(err)
				// TODO I should find another way. mutex is not a good 	solution
				if removeElementFromAckArray(int(part)) {
					mx.Lock()
					*i = *i + 1
					mx.Unlock()
					if *i >= parts {
						udpConn.Close()
					}
				}
			}(&i)
			fmt.Println("i: ", i, "parts: ", parts, "ii: ", ii)
			ii++
		}
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