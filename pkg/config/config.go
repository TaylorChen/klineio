package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

func NewConfig(p string) *viper.Viper {
	envConf := os.Getenv("APP_CONF")
	if envConf == "" {
		envConf = p
	}
	fmt.Println("load conf file:", envConf)
	return getConfig(envConf)
}

func getConfig(path string) *viper.Viper {
	conf := viper.New()
	conf.SetConfigFile(path)
	conf.SetConfigType("yml") // 明确指定配置文件类型
	err := conf.ReadInConfig()
	if err != nil {
		panic(fmt.Sprintf("failed to read config file %s: %v", path, err))
	}
	return conf
}
