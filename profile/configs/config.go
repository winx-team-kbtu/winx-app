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
		GeoIP  geoIP   `yaml:"geoip"`
		Redis  redis   `yaml:"redis"`
		S3     s3      `yaml:"s3"`
		Logger grayLog `yaml:"graylog"`
	}

	app struct {
		Name        string `yaml:"name"`
		Environment string `yaml:"environment"`
		Url         string `yaml:"url"`
		Key         string `yaml:"key"`
	}

	db struct {
		Postgres postgres `yaml:"postgres"`
		Mongo    mongo    `yaml:"mongo"`
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

	mongo struct {
		URI      string `yaml:"uri"`
		Database string `yaml:"database"`
	}

	redis struct {
		Host     string `yaml:"host"`
		Password string `yaml:"password"`
		Port     string `yaml:"port"`
	}

	s3 struct {
		Region          string `yaml:"region"`
		Bucket          string `yaml:"bucket"`
		AccessKeyID     string `yaml:"access_key_id"`
		SecretAccessKey string `yaml:"secret_access_key"`
	}

	geoIP struct {
		BaseURL        string `yaml:"base_url"`
		TimeoutSeconds int    `yaml:"timeout_seconds"`
	}

	grayLog struct {
		Host   string `yaml:"host"`
		Port   int    `yaml:"port"`
		Source string `yaml:"source"`
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
			Mongo: mongo{
				URI:      viper.GetString("db.mongo.uri"),
				Database: viper.GetString("db.mongo.database"),
			},
		},
		GeoIP: geoIP{
			BaseURL:        viper.GetString("geoip.base_url"),
			TimeoutSeconds: viper.GetInt("geoip.timeout_seconds"),
		},
		Redis: redis{
			Host:     viper.GetString("redis.host"),
			Password: viper.GetString("redis.password"),
			Port:     viper.GetString("redis.port"),
		},
		S3: s3{
			Region:          viper.GetString("s3.region"),
			Bucket:          viper.GetString("s3.bucket"),
			AccessKeyID:     viper.GetString("s3.access_key_id"),
			SecretAccessKey: viper.GetString("s3.secret_access_key"),
		},
		Logger: grayLog{
			Host:   viper.GetString("graylog.host"),
			Port:   viper.GetInt("graylog.port"),
			Source: viper.GetString("graylog.source"),
		},
	}
}
