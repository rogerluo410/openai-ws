package server

import (
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

func ProviderRequestMapper(i interface{}, c *Client) interface{} {
	log.Info("开启路由请求映射...")
	return XunfeiVoicedictationRequestParams(i) 
}

func ProviderResponseMapper(i interface{}, c *Client) interface{} {
	log.Info("开启路由响应映射...")
	return XunfeiVoicedictationResponse(i) 
}

func ProviderWsMapper(provider string, apiName string) (*websocket.Conn, error) {
	return XunfeiVoicedictationConn()
}
