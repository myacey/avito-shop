package backconfig

import "github.com/spf13/viper"

type Config struct {
	// 4ALL
	ServerPort   string `mapstructure:"SERVER_PORT"`
	DBPassword   string `mapstructure:"DB_PASSWORD"`
	Testing      bool   `mapstructure:"TESTING"`
	JWTSecretKey string `mapstructure:"JWT_SECRET_KEY"`

	// POSTGRES
	PostgresHost   string `mapstructure:"POSTGRES_HOST"`
	PostgresUser   string `mapstructure:"POSTGRES_USER"`
	PostgresDBName string `mapstructure:"POSTGRES_DB"`
	PostgresPort   string `mapstructure:"POSTGRES_PORT"`

	// REDIS
	RedisUser string `mapstructure:"REDIS_USER"`
	RedisHost string `mapstructure:"REDIS_HOST"`
}

func LoadConfig() (config Config, err error) {
	viper.SetConfigFile(".env")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}
