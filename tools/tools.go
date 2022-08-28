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
package tools

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"strconv"
	"time"
	"github.com/xuyu/goredis"
	"os"
	"log"
	"net/http"
	"strings"
	"io/ioutil"
)
var Clientr *goredis.Redis
var number uint32
type protocolsplit struct {
	Toname []byte
	Mesindex int//message 's index of buffer
}

//[]byte to string
func ByteString(p []byte) string {
	for i := 0; i < len(p); i++ {
		if p[i] == 0 {
			return string(p[0:i])
		}
	}
	return string(p)
}
//get number
func GetNumber() uint32 {
	if number > 4294967290 {
		number = 0
	}
	number = number + 1
	return number
}
//get token
func Gettoken(toname string,key string) string{
	var ti = time.Now().Unix() / 100
	var h = md5.New()
	io.WriteString(h,key +","+toname+","+strconv.FormatInt(ti, 10))
	return hex.EncodeToString(h.Sum(nil))
}
/*
redis client
*/
func Reconnectredis(redis string){
	var err error
	for {
		Clientr, err = goredis.Dial(&goredis.DialConfig{"tcp", redis, 0, "", 10 * time.Second, 10})
		if err != nil {
			log.Println("INFO","redis connect err ",err)
		} else {
			break
		}
		time.Sleep(1 * time.Second)
	}
}
//protocol

func Parsebuffer(c []byte) protocolsplit{
	splitn:=0
	for index := 0; index < len(c); index++ {
		if c[index]=='|' {
			splitn=index
			break
		}		
	}
	var name []byte
	if splitn==0 {
		name=nil
	}else{
		name=c[:splitn]
	}
	return protocolsplit{name,splitn}
}
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func Memset(a []byte, v byte) {
    for i := range a {
        a[i] = v
    }
}
func Array_pos(s []interface{},value interface{}) int {
    for p, v := range s {
        if  v == value  {
            return p
        }
    }
    return -1
}

func HttpPost(url string,post string) ([]byte,error) {
    resp, err := http.Post(url,
        "application/x-www-form-urlencoded",
        strings.NewReader(post))
    if err != nil {
        return []byte{},err
    }
 
    defer resp.Body.Close()
	var body []byte
    body, err = ioutil.ReadAll(resp.Body)
    if err != nil {
        return []byte{},err
    }
    return body,err
}