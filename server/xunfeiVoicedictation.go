// 讯飞平台 - 语音听写
// 文档: https://www.xfyun.cn/doc/asr/voicedictation/API.html#%E6%8E%A5%E5%8F%A3%E8%B0%83%E7%94%A8%E6%B5%81%E7%A8%8B
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
	"reflect"

	"github.com/gorilla/websocket"

	config "github.com/rogerluo410/openai-ws/config"
)

var (
	vdHostUrl   = "wss://iat-api.xfyun.cn/v2/iat"
)

// 发送数据包格式
type Frame struct {
  Data FrameData `json:"data"`
	Common CommonData `json:"common"`
  Business BusinessData `json:"business"`
}

type CommonData struct {
  App_id string `json:"app_id"`
}

type BusinessData struct {
  Language string `json:"language"`
	Domain string `json:"domain"`
	Accent string `json:"accent"`
	Vad_eos int `json:"vad_eos"`
	Dwa string `json:"dwa"`
	Pd string `json:"pd"`
	Ptt int `json:"ptt"`
	Rlang string `json:"rlang"`
	Vinfo int `json:"vinfo"`
	Nunum int `json:"nunum"`
	Speex_size int `json:"speex_size"`
	Nbest int `json:"nbest"`
	Wbest int `json:"wbest"`
}

type FrameData struct {
	Format     string `json:"format"`
	Audio      string `json:"audio"`
	Encoding   string `json:"encoding"`
	Status 	   int `json:"status"`
}

// 接收数据包格式
type RespData struct {
	Sid 	string `json:"sid"`
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    Data  `json:"data"`
}

type Data struct {
	Result Result `json:"result"`
	Status int    `json:"status"`
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

func (w *Ws) String()string  {
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

// 讯飞通用连接
func  XunfeiCommonConn(hostUrlTmp string) (*websocket.Conn, error) {
	// appid     := config.ConfigInstance().XunfeiAppId
  apiSecret := config.ConfigInstance().XunfeiApiSecret
	apiKey    := config.ConfigInstance().XunfeiApiKey

	d := websocket.Dialer{
		HandshakeTimeout: 5 * time.Second,
	}
	//握手并建立websocket连接
	conn, resp, err := d.Dial(assembleAuthUrl(hostUrlTmp, apiKey, apiSecret), nil)
	if err != nil {
		return nil, err
	} else if resp.StatusCode != 101 {
		return nil, err
	}

  return conn, nil
}

// 连接讯飞语音听写的连接串
func XunfeiVoicedictationConn() (*websocket.Conn, error) {
	// appid     := config.ConfigInstance().XunfeiAppId
  apiSecret := config.ConfigInstance().XunfeiApiSecret
	apiKey    := config.ConfigInstance().XunfeiApiKey

	d := websocket.Dialer{
		HandshakeTimeout: 5 * time.Second,
	}
	//握手并建立websocket连接
	conn, resp, err := d.Dial(assembleAuthUrl(vdHostUrl, apiKey, apiSecret), nil)
	if err != nil {
		return nil, err
	} else if resp.StatusCode != 101 {
		return nil, err
	}

  return conn, nil
}

func XunfeiVoicedictationRequestParams(i interface{}) interface{} {
	var frame = Frame{}
	str, ok := i.(string)
	if !ok {
		return nil
	}
	json.Unmarshal([]byte(str), &frame)
	mFrame := ConvertStructToMap(reflect.ValueOf(frame))
	return mFrame
}

// 解析响应串, 暂时弃用, 响应串透传给调用方
func XunfeiVoicedictationResponse(i interface{}) interface{} {
	var resp = RespData{}
	bytes, _ := i.([]byte)
	json.Unmarshal(bytes, &resp)

	return resp.Data.Result.String()
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
