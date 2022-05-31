// 讯飞 - 语音合成
// 文档: https://www.xfyun.cn/doc/tts/online_tts/API.html#%E6%8E%A5%E5%8F%A3%E8%B0%83%E7%94%A8%E6%B5%81%E7%A8%8B
package server

import (
	"encoding/json"
	"reflect"

	"github.com/gorilla/websocket"
)

var (
	ttsHostUrl   = "wss://tts-api.xfyun.cn/v2/tts"
)

type TtsFrame struct {
  Data TtsFrameData `json:"data"`
	Common TtsCommonData `json:"common"`
  Business TtsBusinessData `json:"business"`
}

type TtsCommonData struct {
  App_id string `json:"app_id"`
}

type TtsBusinessData struct {
  Aue string `json:"aue"`
	Sfl int `json:"sfl"`
	Auf string `json:"auf"`
	Vcn string `json:"vcn"`
	Speed int `json:"speed"`
	Volume int `json:"volume"`
	Pitch int `json:"pitch"`
	Bgs int `json:"bgs"`
	Tte string `json:"tte"`
	Reg string `json:"reg"`
	Rdn string `json:"rdn"`
}

type TtsFrameData struct {
	Text     string `json:"text"`
	Status   int `json:"status"`
}

// 连接串
func XunfeiTtsConn() (*websocket.Conn, error) {
	return XunfeiCommonConn(ttsHostUrl)
}

func XunfeiTtsRequestParams(i interface{}) interface{} {
	var frame = TtsFrame{}
	str, ok := i.(string)
	if !ok {
		return nil
	}
	json.Unmarshal([]byte(str), &frame)
	mFrame := ConvertStructToMap(reflect.ValueOf(frame))
	return mFrame
}
