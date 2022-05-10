// 讯飞 - 语音评测
// 文档: https://www.xfyun.cn/doc/Ise/IseAPI.html#%E6%8E%A5%E5%8F%A3%E8%AF%B4%E6%98%8E
package server

import (
	"encoding/json"
	"reflect"

	"github.com/gorilla/websocket"
)

var (
	lseHostUrl   = "wss://ise-api.xfyun.cn/v2/open-ise"
)

type LseFrame struct {
  Data LseFrameData `json:"data"`
	Common LseCommonData `json:"common"`
  Business LseBusinessData `json:"business"`
}

type LseCommonData struct {
  App_id string `json:"app_id"`
}

type LseBusinessData struct {
  Sub string `json:"sub"`
	Ent string `json:"ent"`
	Category string `json:"category"`
	Aus int `json:"aus"`
	Cmd string `json:"cmd"`
	Text string `json:"text"`
	Tte string `json:"tte"`
	Ttp_skip bool `json:"ttp_skip"`
	Extra_ability string `json:"extra_ability"`
	Aue string `json:"aue"`
	Auf string `json:"auf"`
	Rstcd string `json:"rstcd"`
	Group string `json:"group"`
	Check_type string `json:"check_type"`
	Grade string `json:"grade"`
	Rst string `json:"rst"`
	Ise_unite string `json:"ise_unite"`
	Plev string `json:"plev"`
}

type LseFrameData struct {
	Data     string `json:"data"`
	Status   int `json:"status"`
}

// 连接串
func XunfeiLseConn() (*websocket.Conn, error) {
	return XunfeiCommonConn(lseHostUrl)
}

func XunfeiLseRequestParams(i interface{}) interface{} {
	var frame = LseFrame{}
	str, ok := i.(string)
	if !ok {
		return nil
	}
	json.Unmarshal([]byte(str), &frame)
	mFrame := ConvertStructToMap(reflect.ValueOf(frame))
	return mFrame
}
