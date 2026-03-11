package configs

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/viper"
)

var Config *config

type (
	config struct {
		App    app     `yaml:"app"`
		DB     db      `yaml:"db"`
		Redis  redis   `yaml:"redis"`
		Kafka  kafka   `yaml:"kafka"`
		Logger grayLog `yaml:"graylog"`
		Oauth  oauth   `yaml:"oauth"`
	}

	app struct {
		Name        string `yaml:"name"`
		Environment string `yaml:"environment"`
		Url         string `yaml:"url"`
		Key         string `yaml:"key"`
	}

	db struct {
		Postgres postgres `yaml:"postgres"`
	}

	postgres struct {
		Connection string `yaml:"connection"`
		Host       string `yaml:"host"`
		Port       int    `yaml:"port"`
		Database   string `yaml:"database"`
		Username   string `yaml:"username"`
		Password   string `yaml:"password"`
		SSLMode    string `yaml:"sslmode"`
	}

	redis struct {
		Host     string `yaml:"host"`
		Password string `yaml:"password"`
		Port     string `yaml:"port"`
	}

	kafka struct {
		Brokers []string    `yaml:"brokers"`
		Topics  kafkaTopics `yaml:"topics"`
	}

	kafkaTopics struct {
		UserRegistered string `yaml:"user_registered"`
		UserPassword   string `yaml:"user_password"`
	}

	grayLog struct {
		Host   string `yaml:"host"`
		Port   int    `yaml:"port"`
		Source string `yaml:"source"`
	}

	oauth struct {
		AccessTokenExp  int64  `yaml:"access_token_exp"`
		RefreshTokenExp int64  `yaml:"refresh_token_exp"`
		ClientID        string `yaml:"client_id"`
		ClientSecret    string `yaml:"client_secret"`
	}
)

func InitConfig() {
	configName := "config.dev"

	if os.Getenv("APP_ENV") != "" {
		configName += fmt.Sprintf(".%s", os.Getenv("APP_ENV"))
	}

	viper.SetConfigName(configName)
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("../../../..")

	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("не удалось спарсить конфиг файл! Ошибка:%s", err)

		log.Fatal(err)
	}

	Config = &config{
		App: app{
			Name:        viper.GetString("app.name"),
			Environment: viper.GetString("app.environment"),
			Url:         viper.GetString("app.url"),
			Key:         viper.GetString("app.key"),
		},
		DB: db{
			Postgres: postgres{
				Connection: viper.GetString("db.postgres.connection"),
				Host:       viper.GetString("db.postgres.host"),
				Port:       viper.GetInt("db.postgres.port"),
				Database:   viper.GetString("db.postgres.database"),
				Username:   viper.GetString("db.postgres.username"),
				Password:   viper.GetString("db.postgres.password"),
				SSLMode:    viper.GetString("db.postgres.sslmode"),
			},
		},
		Redis: redis{
			Host:     viper.GetString("redis.host"),
			Password: viper.GetString("redis.password"),
			Port:     viper.GetString("redis.port"),
		},
		Kafka: kafka{
			Brokers: viper.GetStringSlice("kafka.brokers"),
			Topics: kafkaTopics{
				UserRegistered: viper.GetString("kafka.topics.user_registered"),
				UserPassword:   viper.GetString("kafka.topics.user_password"),
			},
		},

		Logger: grayLog{
			Host:   viper.GetString("graylog.host"),
			Port:   viper.GetInt("graylog.port"),
			Source: viper.GetString("graylog.source"),
		},
		Oauth: oauth{
			AccessTokenExp:  viper.GetInt64("oauth.access_token_exp"),
			RefreshTokenExp: viper.GetInt64("oauth.refresh_token_exp"),
			ClientID:        "winx-auth-client-id",
			ClientSecret:    "winx-auth-client-secret",
		},
	}
}
