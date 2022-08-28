// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
/*
1、中间断都测试了，没问题
3、压力测试看看性能
*/
package main

import (
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gorilla/websocket"
)

var chann chan bool

func main() {
	i := 1
	b := 0
	n := 1
	for _, v := range os.Args {
		log.Println(v)
	}
	if len(os.Args) < 4 {
		log.Println("执行格式：exec 文件昵称 ws地址 mp4名字（当前目录），")
		return
	}
	chann = make(chan bool)
	for i <= n {
		go upstreamclient("ws://"+os.Args[2]+"/ws/?fnickname="+os.Args[1]+"&fid=yyy"+strconv.Itoa(i)+"&sid=anji123&examtype=examtype111&area=23&format=mp4", i)
		i = i + 1
	}

	for {
		<-chann
		b = b + 1
		if b == n {
			break
		}
	}
}
func abc() {
	log.Println("dddddddd")
}
func upstreamclient(touri string, id int) int {
	var ws *websocket.Conn
	var err error
	ws, _, err = cstDialer.Dial(touri, nil)
	if err != nil {
		chann <- true
		log.Fatal(err)
	}
	log.Println("链接成功")
	c := &connclient{ws: ws}
	c.Init()
	defer func() {
		chann <- true
		log.Println("defer Close")
		c.Close()
	}()
	log.Println("开始上传mp4")
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Println("Abs 已经关闭了", err)
		return 0
	}
	err = c.Sys(dir + "/" + os.Args[3]) //+"/5.mp4"

	if err != nil {
		log.Println("Sys已经关闭了", err)
		return 0
	}
	c.id = id
	go c.Upstream()
	c.Wait()
	log.Println("已经关闭了")
	log.Println("exit process ", id)
	return 0
}
