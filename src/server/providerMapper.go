package server

import (
	log "github.com/sirupsen/logrus"
)

func ProviderMapper(i interface{}) {
	log.Info("开启服务路由映射...")
	if t, ok := i.(*Job); ok {
    log.WithFields(log.Fields{
			"provider": t.Client.Provider,
			"apiName":  t.Client.ApiName,
		}).Info("服务路由映射成功...")
		XunfeiVoicedictationAccess(t) 
	}
  
}