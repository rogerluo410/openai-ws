// 讯飞平台 - 语音合成 - 功能测试
package main

import (
	"fmt"
	"time"
	"encoding/base64"
	"github.com/gorilla/websocket"
	"os"
	"encoding/json"
	"flag"
)
/**
 * 语音合成 
 * 错误码链接：https://www.xfyun.cn/document/error-code （code返回错误码时必看）
 */
var (
	hostUrl   = "ws://localhost:8080/ws"
	queryStr  = "provider=xunfei&api_name=tts"
	token     = "1234"
	appid     = "b55b61a2"
	file      = "./xunfeiTtsTest.mp3" //请填写您的音频文件路径
)

const (
	STATUS_FRAME     = 2  // 数据状态，固定为2
	Text  =  "这是一个语音合成的测试用例"
)

func main() {
	var address string
  flag.StringVar(&address, "a",  hostUrl, "代理ws服务地址")

	var tokenSet string
  flag.StringVar(&tokenSet, "t",  token, "token")

	flag.Parse()

	st:=time.Now()
	d := websocket.Dialer{
		HandshakeTimeout: 5 * time.Second,
	}
	//握手并建立websocket 连接
	url := address+"?"+queryStr+"&token="+tokenSet
	fmt.Println("Connect to ", url)

	conn, resp, err := d.Dial(url, nil)
	if err != nil {
		panic(err.Error())
		return
	}else if resp.StatusCode != 101{
		panic(err.Error())
	}
	defer conn.Close()
	
	frameData := map[string]interface{}{
		"common": map[string]interface{}{
			"app_id": appid, //appid 必须带上，只需第一帧发送
		},
		"business": map[string]interface{}{ //business 参数，只需一帧发送
			"aue": "lame", // raw: pcm格式, lame: mp3格式
			"sfl": 1,
			"vcn": "xiaoyan",
			"tte": "utf8",
		},
		"data": map[string]interface{}{
			"status":    STATUS_FRAME,
			"text":     base64.StdEncoding.EncodeToString([]byte(Text)),
		},
  }
  fmt.Println("数据封装完毕!")

	// 如果音频文件存在, 则删除
	if _, err := os.Stat(file); err == nil {
		e := os.Remove(file)
    if e != nil {
      fmt.Println(e)
    }
	}

	// 发送数据包
	str, _ := json.Marshal(frameData)
	conn.WriteMessage(websocket.TextMessage, str)

	//获取返回的数据
	f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err.Error())
	}

	for {
		fmt.Println("读消息...")
		var resp = RespData{}
		_, msg, err := conn.ReadMessage()
		if err != nil {
				fmt.Println("read message error:", err)
				break
		}
		fmt.Println("读到的消息:" + string(msg))

		json.Unmarshal(msg, &resp)
		// fmt.Println(resp.Data.Audio, resp.Sid)

		// 将audio数据写入文件
		b, err := base64.StdEncoding.DecodeString(resp.Data.Audio)
		fmt.Println(b)
		if err != nil {
      panic(err.Error())
		}
    if _, err := f.Write(b); err != nil {
      panic(err.Error())
    }
  
		if resp.Code != 0 {
			fmt.Println(resp.Code,resp.Message,time.Since(st))
			return
		}
		if resp.Data.Status == 2 {
			fmt.Println(resp.Code,resp.Message, time.Since(st))
			break
		}
	}

	if err := f.Close(); err != nil {
		panic(err)
	}

	time.Sleep(1 * time.Second)
}

type RespData struct {
	Sid 	string `json:"sid"`
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    Data  `json:"data"`
}

type Data struct {
	Audio  string `json:"audio"`
	Ced   string `json:"ced"`
	Status int         `json:"status"`
}

