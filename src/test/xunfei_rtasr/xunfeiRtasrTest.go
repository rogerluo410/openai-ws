// 讯飞平台 - 实时语音转写 - 功能测试
package main

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"golang.org/x/net/websocket"
	"net/url"
	"os"
	"strconv"
	"time"
	log "github.com/sirupsen/logrus"
)

var HOST = "rtasr.xfyun.cn/v1/ws"

var APPID = "123"
var APPKEY = "1234"

// 结束标识
var END_TAG = "{\"end\": true}"

// 每次发送的数据大小
var SLICE_SIZE = 1280

var FILE_PATH = "./xunfeiRtasrTest.pcm"

func main() {
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	mac := hmac.New(sha1.New, []byte(APPKEY))
	strByte := []byte(APPID + ts)
	strMd5Byte := md5.Sum(strByte)
	strMd5 := fmt.Sprintf("%x", strMd5Byte)
	mac.Write([]byte(strMd5))
	signa := url.QueryEscape(base64.StdEncoding.EncodeToString(mac.Sum(nil)))
	requestParam := "appid=" + APPID + "&ts=" + ts + "&signa=" + signa

	url := "ws://" + HOST + "?" + requestParam
  fmt.Println("connect to ", url)
	conn, err := websocket.Dial(url, websocket.SupportedProtocolVersion, "http://" + HOST)
	if err != nil {
		log.Error("err: ", err)
		return
	}

	var message string
	websocket.Message.Receive(conn, &message)
	var m map[string]string
	err = json.Unmarshal([]byte(message), &m)
	println(message)
	if err != nil {
		println(err.Error())
		return
	} else if m["code"] != "0" {
		log.Error("handshake fail!" + message)
		return
	}

	defer conn.Close()
	sendChan := make(chan int, 1)
	readChan := make(chan int, 1)
	defer close(sendChan)
	defer close(readChan)
	go send(conn, sendChan, FILE_PATH)
	go receive(conn, readChan)
	<-sendChan
	<-readChan
}

func send(conn *websocket.Conn, sendChan chan int, filePath string) {
	// 分片上传音频
	defer func() {
		sendChan <- 1
	}()
	data := make([]byte, SLICE_SIZE)
	audio, err := os.Open(filePath)
	defer audio.Close()
	if err != nil {
		log.Error("open file error: " + err.Error())
	}
	for i, e := audio.Read(data); i > 0; i, e = audio.Read(data) {
		if e != nil {
			if e.Error() == "EOF" {
				// 读到结束符
				break
			}
			log.Error("read file error: " + e.Error())
			return
		}
		// println("i:", i)
		if err := websocket.Message.Send(conn, data[:i]); err != nil {
			log.Error("send byte msg err: ", err)
			return
		}
		// println("send data success, sleep 40 ms")
		time.Sleep(40 * time.Millisecond)
	}

	// 上传结束符
	if err := websocket.Message.Send(conn, END_TAG); err != nil {
		log.Error("send string msg err: ", err)
	} else {
		println("send end tag success, ", len(END_TAG))
	}
}

func receive(conn *websocket.Conn, readChan chan int) {
	for {
		var msg []byte
		var result map[string]string
		if err := websocket.Message.Receive(conn, &msg); err != nil {
			if err.Error() == "EOF" {
				println("receive date end")
			} else {
				log.Error("receive msg error: ", err.Error())
			}

			break
		}

		err := json.Unmarshal(msg, &result)
		if err != nil {
			println(string(msg))
			log.Error("response json parse error")
			continue
		}

		if result["code"] == "0" {
			var asrResult AsrResult
			err := json.Unmarshal([]byte(result["data"]), &asrResult)
			if err != nil {
				log.Error("parse asrResult error: " + err.Error())
				println("receive msg: ", string(msg))

				break
			}
			if asrResult.Cn.St.Type == "0" {
				println("------------------------------------------------------------------------------------------------------------------------------------")
				// 最终结果
				for _, wse := range asrResult.Cn.St.Rt[0].Ws {
					for _, cwe := range wse.Cw {
						print(cwe.W)
					}
				}
				println("\r\n------------------------------------------------------------------------------------------------------------------------------------")
			} else {
				for _, wse := range asrResult.Cn.St.Rt[0].Ws {
					for _, cwe := range wse.Cw {
						print(cwe.W)
					}
				}
				println()
			}
		} else {
			println("invalid result: ", string(msg))
		}
	}
	readChan <- 1
}

type AsrResult struct {
	Cn    Cn      `json:"cn"`
	SegId float64 `json:"seg_id"`
}

type Cn struct {
	St St `json:"st"`
}

type St struct {
	Bg   string      `json:"bg"`
	Ed   string      `json:"ed"`
	Type string      `json:"type"`
	Rt   []RtElement `json:"rt"`
}

type RtElement struct {
	Ws []WsElement `json:"ws"`
}

type WsElement struct {
	Wb float64     `json:"wb"`
	We float64     `json:"we"`
	Cw []CwElement `json:"cw"`
}

type CwElement struct {
	W  string `json:"w"`
	Wp string `json:"wp"`
}
