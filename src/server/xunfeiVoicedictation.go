// 讯飞平台 - 语音听写
package server

import (
	"net/url"
	"fmt"
	"time"
	"strings"
	"encoding/base64"
	"crypto/hmac"
	"crypto/sha256"
	"io/ioutil"
	"net/http"
	"encoding/json"

	"github.com/gorilla/websocket"
)

var (
	hostUrl   = "wss://iat-api.xfyun.cn/v2/iat"
  appid     = "b55b61a2"
  apiSecret = "M2FhNGE1ZjE1Nzg1ODQ3MGRkZTkyZWFh"
	apiKey    = "671fc248f8b264ee237a0c29ab624552"	
)

const (
	STATUS_FIRST_FRAME    = 0
	STATUS_CONTINUE_FRAME = 1
	STATUS_LAST_FRAME     = 2
)

func XunfeiVoicedictationAccess(job *Job) {
	fmt.Println(HmacWithShaTobase64("hmac-sha256","hello\nhello","hello"))
	st:= time.Now()
	d := websocket.Dialer{
		HandshakeTimeout: 5 * time.Second,
	}
	//握手并建立websocket连接
	conn, resp, err := d.Dial(assembleAuthUrl(hostUrl, apiKey, apiSecret), nil)
	if err != nil {
			panic(readResp(resp)+err.Error())
			return
	}else if resp.StatusCode !=101{
			panic(readResp(resp)+err.Error())
	}

	if b, err := json.Marshal(job.Message); err != nil {
		panic(err.Error())
		return
	} else {
		conn.WriteJSON(b)
	}

	//获取返回的数据
  for {
		var resp = RespData{}
		_,msg,err := conn.ReadMessage()
		if err != nil {
			fmt.Println("read message error:", err)
			break
		}
		json.Unmarshal(msg, &resp)
		//fmt.Println(string(msg))
		fmt.Println(resp.Data.Result.String(),resp.Sid)

		// 推送给客户端管道
    job.Client.SendMsg <- resp.Data.Result.String()

		if resp.Code != 0 {
			fmt.Println(resp.Code,resp.Message, time.Since(st))
			return
		}
		//decoder.Decode(&resp.Data.Result)
		if resp.Data.Status == 2{
			//cf()
			//fmt.Println("final:",decoder.String())
			fmt.Println(resp.Code,resp.Message, time.Since(st))
			break
			//return
		}
	}
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

//创建鉴权url  apikey 即 hmac username
func assembleAuthUrl(hosturl string, apiKey, apiSecret string) string {
	ul, err := url.Parse(hosturl)
	if err != nil {
			fmt.Println(err)
	}
	//签名时间
	date := time.Now().UTC().Format(time.RFC1123)
	//date = "Tue, 28 May 2019 09:10:42 MST"
	//参与签名的字段 host ,date, request-line
	signString := []string{"host: " + ul.Host, "date: " + date, "GET " + ul.Path + " HTTP/1.1"}
	//拼接签名字符串
	sgin := strings.Join(signString, "\n")
	fmt.Println(sgin)
	//签名结果
	sha := HmacWithShaTobase64("hmac-sha256", sgin, apiSecret)
	fmt.Println(sha)
	//构建请求参数 此时不需要urlencoding
	authUrl := fmt.Sprintf("hmac username=\"%s\", algorithm=\"%s\", headers=\"%s\", signature=\"%s\"", apiKey,
			"hmac-sha256", "host date request-line", sha)
	//将请求参数使用base64编码
	authorization := base64.StdEncoding.EncodeToString([]byte(authUrl))

	v := url.Values{}
	v.Add("host", ul.Host)
	v.Add("date", date)
	v.Add("authorization", authorization)
	//将编码后的字符串url encode后添加到url后面
	callurl := hosturl + "?" + v.Encode()
	return callurl
}

func HmacWithShaTobase64(algorithm, data, key string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(data))
	encodeData := mac.Sum(nil)
	return base64.StdEncoding.EncodeToString(encodeData)
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

//解析返回数据，仅供demo参考，实际场景可能与此不同。
type Decoder struct {
	results []*Result
}

func (d *Decoder)Decode(result *Result)  {
	if len(d.results)<=result.Sn{
		d.results = append(d.results,make([]*Result, result.Sn-len(d.results)+1)...)
	}
	if result.Pgs == "rpl"{
		for i:=result.Rg[0];i<=result.Rg[1];i++{
			d.results[i]=nil
		}
	}
	d.results[result.Sn] = result
}

func (d *Decoder)String()string  {
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

func (t *Result)String() string {
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

func (w *Ws)String()string  {
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