package util

import (
	"time"

	"github.com/spf13/viper"
)

// Config存储应用程序的配置信息
type Config struct {
	RunMode              string        `mapstructure:"RUN_MODE"`
	DBDriver             string        `mapstructure:"DB_DRIVER"`
	DBSource             string        `mapstructure:"DB_SOURCE"`
	MigrationURL         string        `mapstructure:"MIGRATION_URL"`
	TestDBSource         string        `mapstructure:"TEST_DB_SOURCE"`
	HTTPServerAddress    string        `mapstructure:"HTTP_SERVER_ADDRESS"`
	TokenSymmetricKey    string        `mapstructure:"TOKEN_SYMMETRIC_KEY"`
	AccessTokenDuration  time.Duration `mapstructure:"ACCESS_TOKEN_DURATION"`
	RefreshTokenDuration time.Duration `mapstructure:"REFRESH_TOKEN_DURATION"`
	ShiroJID             string        `mapstructure:"SHIRO_JID"`
	PricePerKWh          float64       `mapstructure:"PRICE_PER_KWH"`
	EmailSenderName      string        `mapstructure:"EMAIL_SENDER_NAME"`
	EmailSenderAddress   string        `mapstructure:"EMAIL_SENDER_ADDRESS"`
	EmailSenderPassword  string        `mapstructure:"EMAIL_SENDER_PASSWORD"`
	Username             string        `mapstructure:"USERNAME"`
	Password             string        `mapstructure:"PASSWORD"`
	FullName             string        `mapstructure:"FULL_NAME"`
	Email                string        `mapstructure:"EMAIL"`
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
