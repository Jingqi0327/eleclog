package util

import (
	"github.com/spf13/viper"
)

// Config存储应用程序的配置信息
type Config struct {
	Environment         string  `mapstructure:"ENVIRONMENT"`
	DBDriver            string  `mapstructure:"DB_DRIVER"`
	DBSource            string  `mapstructure:"DB_SOURCE"`
	TestDBSource        string  `mapstructure:"TEST_DB_SOURCE"`
	HTTPServerAddress   string  `mapstructure:"HTTP_SERVER_ADDRESS"`
	ShiroJID            string  `mapstructure:"SHIRO_JID"`
	PricePerKWh         float64 `mapstructure:"PRICE_PER_KWH"`
	EmailSenderName     string  `mapstructure:"EMAIL_SENDER_NAME"`
	EmailSenderAddress  string  `mapstructure:"EMAIL_SENDER_ADDRESS"`
	EmailSenderPassword string  `mapstructure:"EMAIL_SENDER_PASSWORD"`
}

// LoadConfig 从指定路径加载配置文件
func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)  // 指定配置文件路径
	viper.SetConfigName("app") // 指定配置文件名称（不带扩展名）
	viper.SetConfigType("env") // 指定配置文件类型

	viper.AutomaticEnv() // 读取环境变量

	err = viper.ReadInConfig() // 读取配置文件
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config) // 将配置文件映射到Config结构体
	return
}
