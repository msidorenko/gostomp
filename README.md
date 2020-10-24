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
- [x] License

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
    msg := message.New([]byte("Message body"))
    msg.SetDestination("/queue/some_queue")
    
    //Send message to message broker 
    //and wait confirm from server about message accept
    err = client.Producer(msg, gostomp.DELIVERY_SYNC)
    if err != nil {
    	panic(err)
    }

    //Init subscription and define callback function
    subscription := &gostomp.Subscription{
        Destination: "/queue/some_queue",
        Callback: func(msg *message.Message){
            println(string(msg.GetBody()))
            client.Ack(msg)
        },
    }   
    err = client.Subscribe(subscription)
    if err != nil {
        println("ERROR: " + err.Error())
    }
    
    for {}
}
```

## License 
Copyright [2020] Maxim Sidorenko

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.