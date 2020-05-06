package wsserver

import (
	"bytes"
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var addr = flag.String("addr", "localhost:8080", "http service address")

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Client - struct
type Client struct {
	// The websocket connection.
	Conn *websocket.Conn

	// Buffered channel of outbound messages.
	Send    chan []byte
	OutMess Messages
	InMess  Messages
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) readPump() {
	defer func() {
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error { c.Conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))

		// c.Out <- message
		c.InMess.AddMessage(message)
	}
}

// outgoingMessagesRouter func for read messages from array outMess
// and send it in the send of Chan for writePump func
func (c *Client) outgoingMessagesRouter() {
	tick := time.Tick(16 * time.Millisecond)
	for range tick {
		if len(c.OutMess.GetMessages()) > 0 {
			c.Send <- c.OutMess.GetMessages()[0]
			c.OutMess.DelFirstM()
		}
	}
}

// // incomingMessagesRouter func for read messages from out of Chan
// // and add the message to inMess array
// func (c *Client) incomingMessagesRouter() {
// 	for {
// 		c.InMess.AddMessage(<-c.Out)
// 	}
// }

// start - func for start methods of client
func (c *Client) start() {
	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go c.writePump()
	go c.readPump()
	go c.outgoingMessagesRouter()
	// go c.incomingMessagesRouter()
}

// ServeWs handles websocket requests from the peer.
func ServeWs(w http.ResponseWriter, r *http.Request, Clients *Client) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	// add conn to
	(*Clients).Conn = conn
	(*Clients).start()
}

// Start func Start(client *Client)
// start http Websocket server
func Start(Clients *Client) {
	flag.Parse()
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ServeWs(w, r, Clients)
	})
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
