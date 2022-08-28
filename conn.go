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
	//"encoding/json"
	//"io/ioutil"
	"bytes"
	"log"
	"net/url"

	//"net/http"
	//"strconv"
	//"strings"
	//"strconv"
	//"encoding/binary"
	"encoding/binary"
	"encoding/json"
	"os"
	"time"

	//"syscall"
	"upfileserver/tools"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second //Millisecond
	// Time allowed to read the next pong message from the peer.
	pongWait  = 50000 * time.Second
	closeWait = 30 * time.Second
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
type webresp struct {
	Status string
	Info   string
}

//The mount point
var OnConsume []func(message []byte)

// connection is an middleman between the websocket connection and the hub.
type connection struct {
	// The websocket connection.
	ws *websocket.Conn
	// Buffered channel of outbound messages.
	Day       string
	Fname     string
	Fnickname string
	Sid       string
	output    *os.File
	Status    byte
	closed    chan bool
	isclose   bool
	Fid       string
	Examtype  string
	Area      string
	Format    string
	Authurl   string
	Site      string
	Uid       string
}

func (con *connection) Init() {
	con.closed = make(chan bool)
	con.isclose = false
}

// Consume consume messages
func (con *connection) Ack(name string) {
	con.Fname = name
	var err error
	var finfo os.FileInfo
	//oldMask := syscall.Umask(0)
	con.output, err = os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	//syscall.Umask(oldMask)
	if err != nil {
		log.Println("[INFO]", con.Fname, con.Fid, con.Area, con.Examtype, con.Format, con.Sid, con.Fnickname, "conn closed offcnfw ack err ", err)
		con.Close()
		return
	}
	finfo, err = con.output.Stat()
	if err != nil {
		log.Println("[INFO]", con.Fname, con.Fid, con.Area, con.Examtype, con.Format, con.Sid, con.Fnickname, "conn closed offcnfw get stat err ", err)
		con.Close()
		return
	}
	var e [9]byte
	e[0] = byte(0)
	binary.LittleEndian.PutUint64(e[1:], uint64(finfo.Size()))
	log.Println("[INFO]", con.Fname, con.Fid, con.Area, con.Examtype, con.Format, con.Sid, con.Fnickname, "offset size ", finfo.Size())
	t := 0
	for i := 8; i > 1; i-- {
		if e[i] != byte(0) {
			t = i + 1
			break
		}
		t = i
	}

	con.Writebin(e[0:t])
}
func (con *connection) Consume() {
	con.ws.SetReadLimit(maxMessageSize)
	con.ws.SetReadDeadline(time.Now().Add(closeWait))
	var err error
	var message []byte
	var numw int
	var finfo os.FileInfo
	for {
		_, message, err = con.ws.ReadMessage()

		var tmpl int
		if len(message) > 9 {
			tmpl = 9
		} else {
			tmpl = len(message)
		}

		log.Println("[INFO]", con.Fname, con.Fid, con.Area, con.Examtype, con.Format, con.Sid, con.Fnickname, "info receive ", len(message), message[0:tmpl])
		if err != nil {
			log.Println("[INFO]", con.Fname, con.Fid, con.Area, con.Examtype, con.Format, con.Sid, con.Fnickname, "info receive err  ", err)
			break
		}
		if con.isclose {
			continue
		}
		con.ws.SetReadDeadline(time.Now().Add(closeWait))
		if len(message) != 0 {
			c := message[0]
			if c == byte(1) {
				numw, err = con.output.Write(message[1:])
				if err != nil || numw != len(message)-1 {
					log.Println("[ERR]", con.Fname, con.Fid, con.Area, con.Examtype, con.Format, con.Sid, con.Fnickname, "conn closed write to file err ", err)
					break
				}

			} else if c == byte(3) {
				isbreak := true
				if len(message) > 1 && len(message) < 10 {
					finfo, err = con.output.Stat()
					var tmp [8]byte
					for i := 0; i < len(message)-1; i++ {
						tmp[i] = message[i+1]
					}
					sl := int64(binary.LittleEndian.Uint64(tmp[0:]))
					if finfo.Size() != sl {
						log.Println("[ERR]", con.Fname, con.Fid, con.Area, con.Examtype, con.Format, con.Sid, con.Fnickname, "end err the file size is not eq  server size ", finfo.Size(), " client size", sl)
						d := [1]byte{byte(8)}
						err = con.ws.WriteMessage(websocket.BinaryMessage, d[0:])
						if err != nil {
							log.Println("[INFO]", con.Fname, con.Fid, con.Area, con.Examtype, con.Format, con.Sid, con.Fnickname, "conn closed end err ", err)
						}
						con.isclose = true
						continue
					} else {
						//触发切片系统那边的队列
						//fileurl 路径，day 天  filename
						var body []byte
						var b bytes.Buffer
						log.Println("[INFO]", con.Fname, con.Fid, con.Area, con.Examtype, con.Format, con.Sid, con.Fnickname, "posturl  "+con.Authurl+"/?fileurl="+url.QueryEscape(con.Fname)+"&day="+url.QueryEscape(con.Day)+"&sid="+url.QueryEscape(con.Sid)+"&fnickname="+url.QueryEscape(con.Fnickname)+"&fid="+con.Fid+"&examtype="+con.Examtype+"&area="+con.Area+"&site="+con.Site+"&uid="+con.Uid)
						body, err = tools.HttpPost(con.Authurl, "fileurl="+url.QueryEscape(con.Fname)+"&day="+url.QueryEscape(con.Day)+"&sid="+url.QueryEscape(con.Sid)+"&fnickname="+url.QueryEscape(con.Fnickname)+"&fid="+con.Fid+"&examtype="+con.Examtype+"&area="+con.Area+"&site="+con.Site+"&uid="+con.Uid)
						if err != nil {
							b.Reset()

							info := "{\"status\":\"n\",\"info\":\"curl post err\"}"

							var e [5]byte
							e[0] = byte(6)
							bodylen := len(info)
							if bodylen > 1024*1024*4 {
								bodylen = 1024 * 1024 * 4
							}
							binary.LittleEndian.PutUint32(e[1:], uint32(bodylen))
							b.Write(e[0:])
							b.WriteString(info)
							err = con.ws.WriteMessage(websocket.BinaryMessage, b.Bytes())
							log.Println("[ERR]", con.Fname, con.Fid, con.Area, con.Examtype, con.Format, con.Sid, con.Fnickname, "curl post err ", err)
							con.isclose = true
							continue
						}

						webresp := webresp{}
						err := json.Unmarshal(body, &webresp)
						if err != nil {
							b.Reset()

							info := "{\"status\":\"n\",\"info\":\"json decode err\"}"

							var e [5]byte
							e[0] = byte(6)
							bodylen := len(info)
							if bodylen > 1024*1024*4 {
								bodylen = 1024 * 1024 * 4
							}
							binary.LittleEndian.PutUint32(e[1:], uint32(bodylen))
							b.Write(e[0:])
							b.WriteString(info)
							err = con.ws.WriteMessage(websocket.BinaryMessage, b.Bytes())
							log.Println("[ERR]", con.Fname, con.Fid, con.Area, con.Examtype, con.Format, con.Sid, con.Fnickname, "json decode err ", err, "body:", string(body))
							con.isclose = true
							continue
						}
						if webresp.Status != "y" {
							b.Reset()

							var e [5]byte
							e[0] = byte(6)
							bodylen := len(webresp.Info)
							if bodylen > 1024*1024*4 {
								bodylen = 1024 * 1024 * 4
							}
							binary.LittleEndian.PutUint32(e[1:], uint32(bodylen))

							b.Write(e[0:])
							b.WriteString(webresp.Info)
							err = con.ws.WriteMessage(websocket.BinaryMessage, b.Bytes())
							log.Println("[INFO]", con.Fname, con.Fid, con.Area, con.Examtype, con.Format, con.Sid, con.Fnickname, " day=", con.Day, " body=", string(body))
							if err != nil {
								log.Println("[ERR]", con.Fname, con.Fid, con.Area, con.Examtype, con.Format, con.Sid, con.Fnickname, "conn closed end err 6 ", err)
							}
							con.isclose = true
							continue
						}

						d := [1]byte{byte(2)}

						err = con.ws.WriteMessage(websocket.BinaryMessage, d[0:])
						if err != nil {
							log.Println("[ERR]", con.Fname, con.Fid, con.Area, con.Examtype, con.Format, con.Sid, con.Fnickname, "conn closed end err 2 ", err)
						}
						log.Println("[INFO]", con.Fname, con.Fid, con.Area, con.Examtype, con.Format, con.Sid, con.Fnickname, "ok!")

					}
				} else {
					log.Println("[ERR]", con.Fname, con.Fid, con.Area, con.Examtype, con.Format, con.Sid, con.Fnickname, "end err the file size is not exist  ")
					d := [1]byte{byte(8)}
					err = con.ws.WriteMessage(websocket.BinaryMessage, d[0:])
					if err != nil {
						log.Println("[ERR]", con.Fname, con.Fid, con.Area, con.Examtype, con.Format, con.Sid, con.Fnickname, "conn closed end err 8 ", err)
					}
				}
				if isbreak {
					con.isclose = true
					continue
				}
			} else if c == byte(4) {
				err := os.Remove(con.Fname)
				if err != nil {
					log.Println("[ERR] delete ", con.Fname, con.Fid, con.Area, con.Examtype, con.Format, con.Sid, con.Fnickname, err)
				} else {
					log.Println("[INFO] delete ok ", con.Fname, con.Fid, con.Area, con.Examtype, con.Format, con.Sid, con.Fnickname)
				}
				d := [1]byte{byte(9)}
				err = con.ws.WriteMessage(websocket.BinaryMessage, d[0:])
				if err != nil {
					log.Println("[ERR]", con.Fname, con.Fid, con.Area, con.Examtype, con.Format, con.Sid, con.Fnickname, "conn closed end err 9 ", err)
				}
				con.isclose = true
				continue
			} else {
				//其余的垃圾数据一概忽略
			}

		}

	}
	log.Println("send closed", con.Fname, con.Fid, con.Area, con.Examtype, con.Format, con.Sid, con.Fnickname)
	con.closed <- true
}

// Writebin writes a BinaryMessage with the message.
func (con *connection) Writebin(message []byte) error {
	//con.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return con.ws.WriteMessage(websocket.BinaryMessage, message)
}
func (con *connection) Close() error {
	con.output.Close()
	con.ws.Close()
	return nil
}
func (con *connection) Wait() {
	<-con.closed
}
