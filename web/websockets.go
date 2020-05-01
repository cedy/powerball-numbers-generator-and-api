package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"net/http"
	"strings"
	"time"
)

const (
	// Time allowed to write message to the client.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the client.
	pongWait = 60 * time.Second

	// Send pings to client with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

func reader(ws *websocket.Conn) {
	defer ws.Close()
	ws.SetReadLimit(512)
	ws.SetReadDeadline(time.Now().Add(pongWait))
	ws.SetPongHandler(func(string) error { ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			break
		}
	}
}

func writer(ws *websocket.Conn, commBroadcastChan chan chan string) {
	pingTicker := time.NewTicker(pingPeriod)
	// create a channel and append in to the dataChanList, so we can start receiving messages
	dataChan := make(chan string, 100)
	commBroadcastChan <- dataChan
	defer func() {
		commBroadcastChan <- dataChan
		pingTicker.Stop()
		ws.Close()
	}()
	for {
		select {
		case combination := <-dataChan:

			msg := []byte(combination)

			if msg != nil {
				ws.SetWriteDeadline(time.Now().Add(writeWait))
				if err := ws.WriteMessage(websocket.TextMessage, msg); err != nil {
					fmt.Println(err.Error())
					return
				}
			}
		case <-pingTicker.C:
			ws.SetWriteDeadline(time.Now().Add(writeWait))
			if err := ws.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

func serveWs(broadcastCommChan chan chan string) gin.HandlerFunc {
	fn := func(c *gin.Context) {

		upgrader.CheckOrigin = func(r *http.Request) bool {
			// Allow only local connections
			ip := strings.Split(r.RemoteAddr, ":")[0]
			if ip == "127.0.0.1" || ip == "localhost" {
				return true
			}
			return false
		}
		ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			if _, ok := err.(websocket.HandshakeError); !ok {
				fmt.Println(err.Error())
			}
			return
		}

		go writer(ws, broadcastCommChan)
		reader(ws)
	}
	return gin.HandlerFunc(fn)
}
