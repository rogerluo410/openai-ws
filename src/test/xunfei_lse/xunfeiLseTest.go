// 讯飞平台 - 语音评测 - 功能测试
package main

import (
	"fmt"
	"time"
	"encoding/base64"
	"io/ioutil"
	"github.com/gorilla/websocket"
	"os"
	"io"
	"net/http"
	"encoding/json"
	"context"
	"flag"
)
/**
 * 语音评测
 * 错误码链接：https://www.xfyun.cn/document/error-code （code返回错误码时必看）
 */
var (
	hostUrl   = "ws://localhost:8080/ws"
	queryStr  = "provider=xunfei&api_name=lse"
	token     = "1234"
	appid     = "b55b61a2"
	file      = "xunfeiLseTest.pcm" //请填写您的音频文件路径
)

const (
	STATUS_FIRST_FRAME    = 0
	STATUS_CONTINUE_FRAME = 1
	STATUS_LAST_FRAME     = 2
	// BusinessArgs参数常量
	SUB = "ise"
	ENT = "cn_vip"
  // 中文题型：read_syllable（单字朗读，汉语专有）read_word（词语朗读）read_sentence（句子朗读）read_chapter(篇章朗读)
  // 英文题型：read_word（词语朗读）read_sentence（句子朗读）read_chapter(篇章朗读)simple_expression（英文情景反应）read_choice（英文选择题）topic（英文自由题）retell（英文复述题）picture_talk（英文看图说话）oral_translation（英文口头翻译）
  CATEGORY = "read_sentence"
  // 待评测文本 utf8 编码，需要加utf8bom 头
  TEXT = "\uFEFF今天天气怎么样"
  //直接从文件读取的方式
  // TEXT = '\uFEFF'+ open("cn/read_sentence_cn.txt","r",encoding='utf-8').read()
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
		panic(readResp(resp) + err.Error())
		return
	}else if resp.StatusCode != 101{
		panic(readResp(resp) + err.Error())
	}
	//打开音频文件
	var frameSize = 1280              //每一帧的音频大小
	var intervel = 40 * time.Millisecond //发送音频间隔
	//开启协程，发送数据
	ctx,_:=context.WithCancel(context.Background())
	defer conn.Close()
	var status = 0
	go func() {
		//	start:
		audioFile, err := os.Open(file)
		if err != nil {
			panic(err)
		}
		status = STATUS_FIRST_FRAME      //音频的状态信息，标识音频是第一帧，还是中间帧、最后一帧
		var buffer = make([]byte, frameSize)
		for {
			len, err := audioFile.Read(buffer)
			if err != nil {
				if err == io.EOF { //文件读取完了，改变status = STATUS_LAST_FRAME
					status = STATUS_LAST_FRAME
				} else {
					panic(err)
				}
			}
			select {
			case <-ctx.Done():
				fmt.Println("session end ---")
				return
			default:
			}
			switch status {
			case STATUS_FIRST_FRAME: //发送第一帧音频，带business 参数
				frameData := map[string]interface{}{
						"common": map[string]interface{}{
							"app_id": appid, //appid 必须带上，只需第一帧发送
						},
						"business": map[string]interface{}{ //business 参数，只需一帧发送
							"category": CATEGORY, 
							"sub": SUB, 
							"ent": "cn_vip", 
							"cmd": "ssb", 
							"auf": "audio/L16;rate=16000",
              "aue": "raw", 
							"text": TEXT, 
							"ttp_skip": true, 
							"aus": 1,
						},
						"data": map[string]interface{}{
								"status":    STATUS_FIRST_FRAME,
								"text":     base64.StdEncoding.EncodeToString(buffer[:len]),
						},
				}
				fmt.Println("send first")
				// conn.WriteJSON(frameData)
				str, _ := json.Marshal(frameData)
				conn.WriteMessage(websocket.TextMessage, str)
				status = STATUS_CONTINUE_FRAME
			case STATUS_CONTINUE_FRAME:
				frameData := map[string]interface{}{
					  "business":  map[string]interface{}{
							"cmd": "auw", 
							"aus": 2, 
							"aue": "raw",
						},
						"data": map[string]interface{}{
								"status":    STATUS_CONTINUE_FRAME,
								"text":     base64.StdEncoding.EncodeToString(buffer[:len]),
						},
				}
				// conn.WriteJSON(frameData)
				str, _ := json.Marshal(frameData)
				conn.WriteMessage(websocket.TextMessage, str)
			case STATUS_LAST_FRAME:
				frameData := map[string]interface{}{
					"business":  map[string]interface{}{
						"cmd": "auw", 
						"aus": 2, 
						"aue": "raw",
					},
					"data": map[string]interface{}{
							"status":    STATUS_LAST_FRAME,
							"text":     base64.StdEncoding.EncodeToString(buffer[:len]),
					},
				}
				// conn.WriteJSON(frameData)
				str, _ := json.Marshal(frameData)
				conn.WriteMessage(websocket.TextMessage, str)
				fmt.Println("send last")
				return
				//	goto start
			}

			//模拟音频采样间隔
			time.Sleep(intervel)
		}
	}()

	//获取返回的数据
	//var decoder Decoder
	for {
			fmt.Println("读消息...")
			var resp = RespData{}
			_,msg,err := conn.ReadMessage()
			if err != nil {
					fmt.Println("read message error:", err)
					break
			}
			fmt.Println("读到的消息:" + string(msg))

			json.Unmarshal(msg, &resp)
			fmt.Println(resp.Data.Result.String(),resp.Sid)

			if resp.Code!=0{
				fmt.Println(resp.Code,resp.Message,time.Since(st))
				return
			}
			if resp.Data.Status == 2{
				fmt.Println(resp.Code,resp.Message,time.Since(st))
				break
			}
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
    Result Result `json:"result"`
    Status int         `json:"status"`
}

func readResp(resp *http.Response) string {
    if resp == nil {
        return ""
    }
    b, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        panic(err)
    }
    return fmt.Sprintf("code=%d,body=%s", resp.StatusCode, string(b))
}
// 解析返回数据，仅供demo参考，实际场景可能与此不同。
type Decoder struct {
    results []*Result
}

func (d *Decoder) Decode(result *Result) {
    if len(d.results)<=result.Sn{
        d.results = append(d.results,make([]*Result,result.Sn-len(d.results)+1)...)
    }
    if result.Pgs == "rpl"{
        for i:=result.Rg[0];i<=result.Rg[1];i++{
            d.results[i]=nil
        }
    }
    d.results[result.Sn] = result
}

func (d *Decoder) String() string {
	var r string
	for _,v:=range d.results{
			if v== nil{
					continue
			}
			r += v.String()
	}
	return r
}

type Result struct {
	Ls bool `json:"ls"`
	Rg []int `json:"rg"`
	Sn int `json:"sn"`
	Pgs string `json:"pgs"`
	Ws []Ws `json:"ws"`
}

func (t *Result) String() string {
	var wss string
	for _,v:=range t.Ws{
			wss+=v.String()
	}
	return wss
}

type Ws struct {
	Bg int `json:"bg"`
	Cw []Cw `json:"cw"`
}

func (w *Ws) String() string  {
	var wss string
	for _,v:=range w.Cw{
			wss+=v.W
	}
	return wss
}

type Cw struct {
	Sc int `json:"sc"`
	W string `json:"w"`
}