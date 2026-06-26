package util

import (
	"time"

	"github.com/spf13/viper"
)

// Config存储应用程序的配置信息
type Config struct {
	Environment          string        `mapstructure:"ENVIRONMENT"`
	RunMode              string        `mapstructure:"RUN_MODE"`
	DBDriver             string        `mapstructure:"DB_DRIVER"`
	DBSource             string        `mapstructure:"DB_SOURCE"`
	MigrationURL         string        `mapstructure:"MIGRATION_URL"`
	RedisAddress         string        `mapstructure:"REDIS_ADDRESS"`
	HTTPServerAddress    string        `mapstructure:"HTTP_SERVER_ADDRESS"`
	GrpcServerAddress    string        `mapstructure:"GRPC_SERVER_ADDRESS"`
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
	DetectLowBalanceCron string        `mapstructure:"DETECT_LOW_BALANCE_CRON"`
	FetchSurplusCron     string        `mapstructure:"FETCH_SURPLUS_CRON"`
	FrontendOrigin       string        `mapstructure:"FRONTEND_ORIGIN"`
	ProxygRPCAddress     string        `mapstructure:"PROXY_GRPC_ADDRESS"`
	RedisLimiterCapacity int           `mapstructure:"REDIS_LIMITER_CAPACITY"`
	RedisLimiterRate     float64       `mapstructure:"REDIS_LIMITER_RATE"`
}

// LoadConfig 从指定路径加载配置文件
func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)  // 指定配置文件路径
	viper.SetConfigName("app") // 指定配置文件名称（不带扩展名）
	viper.SetConfigType("env") // 指定配置文件类型

	viper.AutomaticEnv() // 读取环境变量

	// 手动绑定环境变量键名，否则在没有配置文件的情况下 Unmarshal 无法知道要读取哪些环境变量
	envKeys := []string{
		"ENVIRONMENT", "RUN_MODE", "DB_DRIVER", "DB_SOURCE", "MIGRATION_URL",
		"REDIS_ADDRESS", "HTTP_SERVER_ADDRESS", "GRPC_SERVER_ADDRESS",
		"TOKEN_SYMMETRIC_KEY", "ACCESS_TOKEN_DURATION", "REFRESH_TOKEN_DURATION",
		"SHIRO_JID", "PRICE_PER_KWH", "EMAIL_SENDER_NAME", "EMAIL_SENDER_ADDRESS",
		"EMAIL_SENDER_PASSWORD", "USERNAME", "PASSWORD", "FULL_NAME", "EMAIL",
		"DETECT_LOW_BALANCE_CRON", "FETCH_SURPLUS_CRON", "FRONTEND_ORIGIN",
		"PROXY_GRPC_ADDRESS", "REDIS_LIMITER_CAPACITY", "REDIS_LIMITER_RATE",
	}
	for _, k := range envKeys {
		viper.BindEnv(k)
	}

	err = viper.ReadInConfig() // 读取配置文件
	if err != nil {
		// 如果是因为找不到配置文件导致的错误，我们忽略它（因为生产环境下我们全靠 K8s 环境变量）
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return
		}
	}

	err = viper.Unmarshal(&config) // 将配置文件映射到Config结构体
	return
}
