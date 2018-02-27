package main

import (
	"udp-go/client"
)


const (
	Host = "89.42.211.10"
	Port = "1313"
	ThreadNumbers = 5
)

func main() {
	client.InitClient(Host, Port, ThreadNumbers)
	//client.Send([]byte("asdfadsfadsfadsf"))
}