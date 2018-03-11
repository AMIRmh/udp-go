package main

import (
	"../src/client"

	"sync"
	"fmt"
)


const (
	Host = "89.42.211.10"
	Port = "1313"
	ThreadNumbers = 5
)

func main() {
	client.InitClient(Host, Port, ThreadNumbers)
	//client.Send([]byte("asdfadsfadsfadsf"))
	//reader := bufio.NewReader(os.Stdin)
	//reader.ReadString('\n')
	c := make(chan int, 200000000)
	var wg sync.WaitGroup
	go func() {

		for i := 0; i < 200000000; i++ {
			wg.Add(1)
			go test(c, i, &wg)
		}
		wg.Wait()
		close(c)
	}()

	x := 0
	for range c {
		x++
	}
	fmt.Println(x)


}


func test(c chan int, i int, wg *sync.WaitGroup) {
	c <- i
	wg.Done()
}