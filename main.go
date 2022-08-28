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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	//"time"
	"sync"

	"strconv"
	"upfileserver/tools"
)

/**
 */

var Configurationptr map[string]interface{}
var number uint32
var locknum sync.Mutex
var syspath string
var Authurl string
var conmanage map[string]*connection
var lockconm sync.Mutex

func Getconfig(name string) interface{} {
	v, ok := Configurationptr[name]
	if ok {
		return v
	}
	return nil
}
func logjoin(fname, fid, sid, examtype, area, format string) string {
	return "fname=" + fname + " " + "fid=" + fid + " " + "sid=" + sid + " " + "examtype=" + examtype + " " + "area=" + area
}
func main() {
	//runtime.GOMAXPROCS(8)

	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal("[ERR]", err)
	}
	syspath = dir
	//syspath="/home/offcn/mygo/upfileserver"
	if len(os.Args) != 2 {
		log.Fatal("[ERR]", "type:exe conf.json")
	}

	isexits, _ := tools.PathExists(syspath + "/log/")
	if !isexits {
		isexits, _ := tools.PathExists(syspath + "/log/")
		if !isexits {
			//oldMask := syscall.Umask(0)
			err = os.Mkdir(syspath+"/log/", 0777)
			//syscall.Umask(oldMask)
			if err != nil {
				log.Fatal("[ERR]", err)
			}
		}
	}
	var f *os.File
	f, err = os.OpenFile(syspath+"/log/access.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		log.Fatal("[ERR]", err)
	}
	log.SetOutput(f)

	filebytes, _ := ioutil.ReadFile(os.Args[1])
	err = json.Unmarshal(filebytes, &Configurationptr)

	if err != nil {
		log.Fatal("[ERR]", "conf.json:", err)
	}
	//go h.run()
	conmanage = make(map[string]*connection)

	authurl_i := Getconfig("authurl")
	if authurl_i == nil {
		log.Fatal("[ERR]", "main", "socket not found")
	}
	var ok bool
	Authurl, ok = authurl_i.(string)
	if !ok || Authurl == "" {
		log.Fatal("[ERR]", "main", "Authurl not found")
	}

	socket_i := Getconfig("socket")

	if socket_i == nil {
		log.Fatal("[ERR]", "main", "socket not found")
	}
	socket, ok := socket_i.(string)
	if !ok || socket == "" {
		log.Fatal("[ERR]", "main", "socket not found")
	}
	tcpAddr, _ := net.ResolveTCPAddr("tcp4", socket)
	listener, _ := net.ListenTCP("tcp4", tcpAddr)
	http.HandleFunc("/ws/", serveWs)
	http.HandleFunc("/httpadmin/", serveHttp)
	log.Println("[INFO]", "sever is started")
	err = http.Serve(listener, nil)
	if err != nil {
		log.Fatal("[ERR]", "ListenAndServe: ", err)
	}
}
func serveHttp(w http.ResponseWriter, r *http.Request) {

}

// serverWs handles websocket requests from the peer.
func serveWs(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "upgrade err ", 405)
		log.Println("[INFO]", err)
		return
	}
	//Link and log in
	err = r.ParseForm()
	if err != nil {
		http.Error(w, "httpform parse err", 405)
		log.Println("[INFO]", "main", "httpform parse ", err)
		return
	}
	//access the callback unfinished
	day := strconv.Itoa(time.Now().Day())
	//此处的安全类似浏览器的级别，只校验sid
	fid := r.FormValue("fid") //唯一标识视频的id
	sid := r.FormValue("sid")
	examtype := r.FormValue("examtype")
	area := r.FormValue("area")
	format := r.FormValue("format")
	fnickname := r.FormValue("fnickname")
	site := r.FormValue("site")
	uid := r.FormValue("uid")
	if fnickname == "" || fid == "" || len(fid) < 5 || sid == "" || examtype == "" || area == "" || format == "" || uid == "" {
		http.Error(w, "新来的链接fnickname，fid，sid，examtype，area,format,uid不能为空", 405)
		log.Println("[ERR]", "新来的链接fnickname，fid，sid，examtype，area,format,uid不能为空", r.Form)
		return
	}
	//鉴权
	hc := http.Client{}
	r.Form.Add("sid", sid)

	authrest_i := Getconfig("authrest")
	if authrest_i == nil {
		http.Error(w, "authrest not found", 405)
		log.Fatal("[ERR]", "main", "authrest not found", logjoin(fnickname, fid, sid, examtype, area, format))
	}
	authrest, ok := authrest_i.(string)
	if !ok || authrest == "" {
		http.Error(w, "authrest not found", 405)
		log.Fatal("[ERR]", "main", "authrest not found", logjoin(fnickname, fid, sid, examtype, area, format))
	}

	req, _ := http.NewRequest("POST", authrest, strings.NewReader(r.Form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err0 := hc.Do(req)
	if err0 != nil {
		http.Error(w, "鉴权接口访问失败", 405)
		log.Println("[ERR]", "鉴权接口访问失败", logjoin(fnickname, fid, sid, examtype, area, format), err0)
		return
	}
	body, err1 := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if string(body) != "1" || err1 != nil {
		http.Error(w, "鉴权错误", 405)
		log.Println("[ERR]", "main", "鉴权错误", logjoin(fnickname, fid, sid, examtype, area, format), string(body))
	}
	var c *connection
	c = &connection{ws: ws, Day: day, Sid: sid, Fname: "", Fid: fid, Examtype: examtype, Area: area, Authurl: Authurl, Fnickname: fnickname, Site: site, Uid: uid}
	c.Init()
	lockconm.Lock()
	_, ok = conmanage[day+"_"+fid]
	lockconm.Unlock()
	if ok == true {
		var e [1]byte
		e[0] = byte(4)
		c.Writebin(e[0:])
		log.Println("[ERR]", "新来的链接名称重复", logjoin(fnickname, fid, sid, examtype, area, format))
		c.Close()
		return
	}
	defer func() {
		c.Close()
		log.Println("defer", day, fnickname)
		lockconm.Lock()
		delete(conmanage, day+"_"+fnickname)
		lockconm.Unlock()
	}()

	path_i := Getconfig("path")

	if path_i == nil {
		log.Fatal("[ERR]", "main", "socket not found")
	}
	path, ok := path_i.(string)
	if !ok || path == "" {
		log.Fatal("[ERR]", "main", "socket not found")
	}

	fldir := fmt.Sprintf("%s%s/", path, fid[0:2])
	// oldMask := syscall.Umask(0)
	os.MkdirAll(fldir, 0777)
	//syscall.Umask(oldMask)
	flname := fldir + fid + "." + format
	log.Println("[INFO]", "New links came in", flname)
	c.Ack(flname)
	go c.Consume()
	lockconm.Lock()
	conmanage[c.Fid] = c
	lockconm.Unlock()
	c.Wait()
	log.Println("[INFO]", "链接关闭")
}
