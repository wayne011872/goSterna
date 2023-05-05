package golanggeneral

import (
	"os"
	"github.com/wayne011872/goSterna/util"
	"github.com/spf13/viper"
)

const (
	CtxServDiKey = util.CtxKey("ServiceDI")
)

func InitConfByEnv(di interface{}) {
	configPath := os.Getenv(("CONFIG_PATH"))
	configName := os.Getenv(("CONFIG_NAME"))
	configType := os.Getenv(("CONFIG_TYPE"))
	if configPath == "" {
		panic("沒有設定CONFIG_PATH參數")
	}
	if configName == "" {
		panic("沒有設定CONFIG_NAME參數")
	}
	if configType == "" {
		panic("沒有設定CONFIG_TYPE參數")
	}
	vip := viper.New()
	vip.AddConfigPath(configPath)
	vip.SetConfigName(configName)
	vip.SetConfigType(configType)
	if err := vip.ReadInConfig(); err != nil {
		panic(err)
	}
	err := vip.UnmarshalKey("mongo", &di)
	if err != nil {
		panic(err)
	}
}