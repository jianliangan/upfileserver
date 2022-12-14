/*
Copyright 2015 anzi

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package main

import (
	"log"

	"encoding/binary"
	"io"
	"os"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second //Millisecond
	// Time allowed to read the next pong message from the peer.
	pongWait = 50 * time.Second
	// Send pings to peer with this period. Must be less than pongWait.
	//pingPeriod = (pongWait * 9) / 10
	// Maximum message size allowed from peer.
	maxMessageSize = 1000 * 1024 * 1024
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type messageb struct {
	uc  string
	msg []byte
}

//The mount point
var OnConsume []func(message []byte)

// connclient is an middleman between the websocket connclient and the hub.
type connclient struct {
	// The websocket connclient.
	ws *websocket.Conn
	// Buffered channel of outbound messages.
	day    string
	fname  string
	output *os.File
	status byte
	id     int
	closed chan bool
	offset int64
}

var cstDialer = websocket.Dialer{
	Subprotocols:    []string{},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func (con *connclient) Init() {
	con.closed = make(chan bool)
}

// Consume consume messages
func (con *connclient) Sys(name string) error {
	con.fname = name
	var err error
	con.output, err = os.OpenFile(name, os.O_RDWR, 0666)
	if err != nil {
		log.Println("[INFO]", "1conn closed offcnfw ack err ", err)
		return err
	}
	log.Println("æå¼mp4æå")
	return nil
}
func (con *connclient) Consume() {
	con.ws.SetReadLimit(maxMessageSize)
	var err error
	var message []byte
	for {
		log.Println("å¼å§æ¥åä¿¡æ¯")
		_, message, err = con.ws.ReadMessage()
		log.Println("æ¶å°æ°æ®", message)
		if err != nil {
			log.Println("çå¬éè¯¯", err)
			break
		}
		if len(message) != 0 {
			c := message[0]
			if c == byte(0) {
				if con.status == byte(1) {
					continue
				}
								var tmp [8]byte
				for i := 0; i < len(message)-1; i++ {         		 
						tmp[i]=message[i+1]				 
    			}
				var len uint64
				len = binary.LittleEndian.Uint64(tmp[0:])
				con.offset = int64(len)
				con.status = byte(1)
				go con.Upstream()
				log.Println("æ¡æä¿¡æ¯")
			}
			if c == byte(2) {
				log.Println("åè®®å®æ¯")

				break
			}
			if c == byte(4) {
				log.Println("æµéå¤äºç»æ")

				log.Println("3333")
				break
			}
			if c == byte(8) {
				log.Println("æµç»ææ¶ï¼é¿åº¦ä¸å¯¹")

				log.Println("3333")
				break
			}
			if c == byte(8) {
				log.Println("æä»¶éåçæ¶åæ¥éäº")

				log.Println("3333")
				break
			}
		}
	}
	con.closed <- true
}

func (con *connclient) Upstream() {
	buf := make([]byte, 1024)
	log.Println("ä¸ä¼ æµ")
f:
	for {
		num, err := con.output.ReadAt(buf[1:], con.offset)
		//log.Println("num ", err,num)
		if err != nil {
			if err == io.EOF {
				if num == 0 {
					//æ°å¢
					buf[0] = byte(3)
					log.Println("client send end to server")
					binary.LittleEndian.PutUint64(buf[1:], uint64(con.offset))
					err = con.Writebin(buf[0 : 9])
					if err != nil {
						log.Println("[ERR] ", err)
						break f
					}
					break
					//ç»æ
					time.Sleep(time.Second * 1)
					continue
				}
			} else {
				log.Println("[ERR] è¯»æä»¶éè¯¯", err)
				break f
			}
		}
		buf[0] = byte(1)
		err = con.Writebin(buf[0 : num+1])
		if err != nil {
			log.Println("[ERR] ", err)
			break f
		}
		con.offset = con.offset + int64(num)
		// log.Println("ééå°ï¼",num,"å­è"," åç§»éï¼",con.offset)
	}
}

// Writebin writes a BinaryMessage with the message.
func (con *connclient) Writebin(message []byte) error {
	///con.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return con.ws.WriteMessage(websocket.BinaryMessage, message)
}
func (con *connclient) Close() error {
	log.Println("[INFO] close the socket")
	con.output.Close()
	con.ws.Close()
	return nil
}
func (con *connclient) Wait() {
	<-con.closed
}
