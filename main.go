package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
	ws "wsserver/wsserver"
)

// Message - test
type Message struct {
	Rout    string `json:"r"`
	Message string `json:"m"`
}

func main() {
	fmt.Println("start server")
	var Client ws.Client = ws.Client{
		Send:    make(chan []byte, 512),
		OutMess: make(ws.Messages, 0, 200),
		InMess:  make(ws.Messages, 0, 200),
	}
	go ws.Start(&Client)

	var mes Message
	tick := time.Tick(16 * time.Millisecond)
	for range tick {

		if len(Client.InMess.GetMessages()) > 0 {
			err := json.Unmarshal(Client.InMess.GetMessages()[0], &mes)
			if err != nil {
				log.Println(err)
			}
			switch mes.Rout {
			case "test1":
				jsonData, err := json.Marshal(Message{"ready", "gas"})
				if err != nil {
					log.Printf("error: %v", err)
					break
				}
				Client.OutMess.AddMessage(jsonData)
				Client.InMess.DelFirstM()
			case "test2":
				for i := 0; i < 10; i++ {
					m := string(i) + "gas"
					jsonData, err := json.Marshal(Message{"test2-ready", m})
					if err != nil {
						log.Printf("error: %v", err)
						break
					}
					Client.OutMess.AddMessage(jsonData)

				}
				Client.InMess.DelFirstM()
			default:
				fmt.Println("message incorrectly")
				Client.InMess.DelFirstM()
			}
		}
	}
}
