// 讯飞平台 - 语音听写 - 功能测试
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
 * 语音听写流式 WebAPI 接口调用示例 接口文档（必看）：https://doc.xfyun.cn/rest_api/语音听写（流式版）.html
 * webapi 听写服务参考帖子（必看）：http://bbs.xfyun.cn/forum.php?mod=viewthread&tid=38947&extra=
 * 语音听写流式WebAPI 服务，热词使用方式：登陆开放平台https://www.xfyun.cn/后，找到控制台--我的应用---语音听写---服务管理--上传热词
 * 注意：热词只能在识别的时候会增加热词的识别权重，需要注意的是增加相应词条的识别率，但并不是绝对的，具体效果以您测试为准。
 * 错误码链接：https://www.xfyun.cn/document/error-code （code返回错误码时必看）
 * @author iflytek
 */
var (
	hostUrl   = "ws://localhost:8080/ws"
	queryStr  = "provider=xunfei&api_name=voicedictation"
	token     = "1234"
	appid     = "b55b61a2"
	file      = "xufeiVoiceDictationTest_16k_10.pcm" //请填写您的音频文件路径
)

const (
	STATUS_FIRST_FRAME    = 0
	STATUS_CONTINUE_FRAME = 1
	STATUS_LAST_FRAME     = 2
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
								"language":"zh_cn",
								"domain":"iat",
								"accent":"mandarin",
						},
						"data": map[string]interface{}{
								"status":    STATUS_FIRST_FRAME,
								"format":    "audio/L16;rate=16000",
								"audio":     base64.StdEncoding.EncodeToString(buffer[:len]),
								"encoding":  "raw",
						},
				}
				fmt.Println("send first")
				// conn.WriteJSON(frameData)
				str, _ := json.Marshal(frameData)
				conn.WriteMessage(websocket.TextMessage, str)
				status = STATUS_CONTINUE_FRAME
			case STATUS_CONTINUE_FRAME:
				frameData := map[string]interface{}{
						"data": map[string]interface{}{
								"status":    STATUS_CONTINUE_FRAME,
								"format":    "audio/L16;rate=16000",
								"audio":     base64.StdEncoding.EncodeToString(buffer[:len]),
								"encoding":  "raw",
						},
				}
				// conn.WriteJSON(frameData)
				str, _ := json.Marshal(frameData)
				conn.WriteMessage(websocket.TextMessage, str)
			case STATUS_LAST_FRAME:
				frameData := map[string]interface{}{
						"data": map[string]interface{}{
								"status":    STATUS_LAST_FRAME,
								"format":    "audio/L16;rate=16000",
								"audio":     base64.StdEncoding.EncodeToString(buffer[:len]),
								"encoding":  "raw",
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