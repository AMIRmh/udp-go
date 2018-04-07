# Reliable UDP

## The idea

the idea is very easy! send the packet and wait for the ack. but obviously with threads!!!

## The challenges

As UDP does not have any response, it means that the packets are just sent and the sender is not waiting for the response.
so this should be handled by the code itself.
Since Go provides a perfect and easy multi threading solutions, debugging gets easier but still asynchronous code has bugs
potentially.

## How to use

Well up to now, this is just a POC and can't be used in a real project. client is already finished and just server needs
some minor developments.

### Client usage

Just pass the server address, port and number of threads to `InitClient` function and then call `Send` function filled
with your data.

```go
import "udp-go/Workspace/src/client"

func main() {
    client.InitClient(Host, Port, NumberOfThreads)
    client.Send([]byte("Hello, World!!!"))
}

```