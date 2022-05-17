package config

import (
	"os"
  "io/ioutil"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

const CONFIG_FILE = "config.yml"

var configVar *config

func init() {
	configVar = &config{}
	configVar.readConfig()
}

func ConfigInstance() *config {
  return configVar
}

type config struct {
	YituDevId string `yaml:"yitu_dev_id"`
	YituDevKey string `yaml:"yitu_dev_key"`
	XunfeiAppId string `yaml:"xunfei_app_id"`
	XunfeiApiSecret string `yaml:"xunfei_api_secret"`
	XunfeiApiKey string `yaml:"xunfei_api_key"`
}

func (c *config) readConfig() {
	// Open YAML file
	// Get dir path of exec file
	dirPath, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.WithField("Err", err).Fatal("Can not find exec dir path")
	}

	// 配置文件和可执行程序放入同一目录下
	filePath := filepath.Join(dirPath, CONFIG_FILE)

	// ioutil.ReadFile 读取整个文件
	// os.Read 从文件中读缓冲区大小的数据
	buf, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.WithField("Err", err).Fatalf("Can not find %s", filePath)
	}

	err = yaml.Unmarshal(buf, c)
	if err != nil {
		log.WithField("Err", err).Fatal("Read config.yml failed")
	}

	log.WithField("Config", c).Info("Read config.yml")
}