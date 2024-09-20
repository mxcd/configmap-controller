package util

import "github.com/mxcd/go-config/config"

func InitConfig() error {
	err := config.LoadConfigWithOptions([]config.Value{
		config.String("LOG_LEVEL").NotEmpty().Default("info"),
		config.Bool("DEV").Default(false),

		config.String("REDIS_HOST").NotEmpty().Default("localhost"),
		config.Int("REDIS_PORT").Default(6379),
		config.String("REDIS_PASSWORD").Default(""),
		config.Int("REDIS_DATABASE_INDEX").Default(0),
		config.Bool("REDIS_SENTINEL").Default(false),
	}, &config.LoadConfigOptions{
		DotEnvFile: "controller.env",
	})
	return err
}
