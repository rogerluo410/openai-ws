package server

import (
	"fmt"
	"errors"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

var keys = []string{"序号", "服务名", "服务名中文说明", "API名", "API名中文说明", "通信协议"}
var dict = []map[string]string{
	map[string]string{"序号": "1", "服务名": "xunfei", "服务名中文说明": "讯飞", "API名": "voicedictation", "API名中文说明": "语音听写", "通信协议": "ws"},
	map[string]string{"序号": "2", "服务名": "xunfei", "服务名中文说明": "讯飞", "API名": "tts", "API名中文说明": "语音合成", "通信协议": "ws"},
	map[string]string{"序号": "3", "服务名": "xunfei", "服务名中文说明": "讯飞", "API名": "rtasr", "API名中文说明": "实时语音转写", "通信协议": "ws"},
	map[string]string{"序号": "4", "服务名": "xunfei", "服务名中文说明": "讯飞", "API名": "lse", "API名中文说明": "语音评测", "通信协议": "ws"},
} 

func MapDict() {
  for _, d := range dict {
    for _, k :=range keys {
			fmt.Printf("%s: %s  |  ", k, d[k])
		}
		fmt.Println("\n")
	}
}

func ProviderRequestMapper(i interface{}, c *Client) interface{} {
	log.Info("查询路由请求映射...")
  
	switch c.Provider {
	case "xunfei": 
	  switch c.ApiName {
		case "voicedictation":
			return XunfeiVoicedictationRequestParams(i)
		case "tts":
			return XunfeiTtsRequestParams(i)
		case "lse":
			return XunfeiLseRequestParams(i)
		case "rtasr":
			return XunfeiRtasrRequestParams(i)
		default:
			return nil
		}
	default:
		return nil
	}
	
}

func ProviderResponseMapper(i interface{}, c *Client) interface{} {
	log.Info("查询路由响应映射...")

	switch c.Provider {
	case "xunfei": 
		switch c.ApiName {
		case "voicedictation":
			return XunfeiVoicedictationResponse(i)
		default:
			return nil
		}
	default:
		return nil
	}
}

func ProviderWsMapper(provider string, apiName string) (*websocket.Conn, error) {
	log.Info("查询云服务连接串...")

	switch provider {
		case "xunfei": 
			switch apiName {
			case "voicedictation":
				return XunfeiVoicedictationConn()
			case "tts":
				return XunfeiTtsConn()
			case "lse":
				return XunfeiLseConn()
			case "rtasr":
				return XunfeiRtasrConn()
			default:
				return nil, errors.New("讯飞服务未知的API名")
			}
		default:
			return nil, errors.New("未知服务类型")
		}
}
