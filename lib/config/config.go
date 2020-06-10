package config

import (
	"fmt"
	"log"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// Config 配置文件
type Config struct {
	Host string
	Port string
}

// Init 初始化函数
func Init() error {
	c := Config{}
	// 初始化配置文件
	if err := c.initConfig(); err != nil {
		return err
	}

	c.watchConfig()

	return nil
}

// initConfig aa
func (c *Config) initConfig() error {
	// v := viper.New()
	viper.AddConfigPath("$GOPATH/src/pigeon/conf")
	viper.SetConfigName("pigeon")

	// 设置配置文件格式为toml
	viper.SetConfigType("toml")
	// viper解析配置文件

	if err := viper.ReadInConfig(); err != nil {
		log.Println(err)
		return err
	}

	return nil
}

// 监听配置文件是否改变,用于热更新
func (c *Config) watchConfig() {
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		fmt.Printf("Config file changed: %s\n", e.Name)
		c.Host = viper.GetString("base.ip")
		c.Port = viper.GetString("base.port")
	})
}
