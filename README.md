# Golang stomp package

Just another implementation of a STOMP client library :-)

Inspired by https://github.com/go-stomp/stomp 

### Todo
- [x] Producer sync mode
- [ ] Heart-beat  
- [ ] Catch message broker errors
- [ ] SSL support
- [ ] Examples for all cases
- [ ] High level library: Topic/Queue  
- [ ] Tests
- [ ] Documentation
- [ ] License

## Usage Instructions

```
go get github.com/msidorenko/gostomp
```

```go
package main
import (
	"github.com/msidorenko/gostomp"
	"github.com/msidorenko/gostomp/message"
)

func main(){
    client, err := gostomp.NewClient("tcp://localhost:61613")
    if err != nil {
    	panic(err)
    }

    err = client.Connect()
    if err != nil {
        panic(err)
    }

    //Create new message
    msg := message.New("Message body")
    msg.SetDestination("/queue/some_queue")
    
    //Send message to message broker 
    //and wait confirm from server about message accept
    client.Producer(msg, gostomp.DELIVERY_SYNC)

    //Init subscription and define callback function
    subscription := &gostomp.Subscription{
        Destination: "/queue/some_queue",
        Callback: func(msg *message.Message){
            println(msg.GetBody())
            client.Ack(msg)
        },
    }
    err = client.Subscribe(subscription)
    if err != nil {
        println("ERROR: " + err.Error())
    }
}
```

## License
Later. But this library always be free for usage.
