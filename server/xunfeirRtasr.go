// 讯飞 - 实时语音转写
// 文档: https://www.xfyun.cn/doc/asr/rtasr/API.html#%E6%8E%A5%E5%8F%A3%E8%AF%B4%E6%98%8E
package server

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/gorilla/websocket"

	config "github.com/rogerluo410/openai-ws/config"
)

var (
	rtasrHostUrl   = "wss://rtasr.xfyun.cn/v1/ws"
)

// 连接串
func XunfeiRtasrConn() (*websocket.Conn, error) {
	appid     := config.ConfigInstance().XunfeiAppId
  // apiSecret := config.ConfigInstance().XunfeiApiSecret
	apiKey    := config.ConfigInstance().XunfeiApiKey

	ts := strconv.FormatInt(time.Now().Unix(), 10)
	mac := hmac.New(sha1.New, []byte(apiKey))
	strByte := []byte(appid + ts)
	strMd5Byte := md5.Sum(strByte)
	strMd5 := fmt.Sprintf("%x", strMd5Byte)
	mac.Write([]byte(strMd5))
	signa := url.QueryEscape(base64.StdEncoding.EncodeToString(mac.Sum(nil)))
	requestParam := "appid=" + appid + "&ts=" + ts + "&signa=" + signa

	d := websocket.Dialer{
		HandshakeTimeout: 5 * time.Second,
	}
	//握手并建立websocket连接
	conn, resp, err := d.Dial(rtasrHostUrl + "?" + requestParam, nil)
	if err != nil {
		return nil, err
	} else if resp.StatusCode != 101 {
		return nil, err
	}

  return conn, nil
}

func XunfeiRtasrRequestParams(i interface{}) interface{} {
  str, ok := i.(string)
	if !ok {
		return nil
	}
	return str
}