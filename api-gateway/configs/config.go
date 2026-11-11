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
		App      app      `yaml:"app"`
		Services services `yaml:"services"`
		Swagger  swagger  `yaml:"swagger"`
		Pilot    pilot    `yaml:"pilot"`
	}

	pilot struct {
		// Mode disables the legacy Social features (photos, swipes, matches,
		// recommendations, chat) at the gateway. It defaults to true so a missing
		// setting never exposes Social features to pilot users.
		Mode bool `yaml:"mode"`
	}

	app struct {
		Name        string `yaml:"name"`
		Environment string `yaml:"environment"`
		Url         string `yaml:"url"`
		Key         string `yaml:"key"`
	}

	services struct {
		Auth           service `yaml:"auth"`
		Notification   service `yaml:"notification"`
		Profile        service `yaml:"profile"`
		Match          service `yaml:"match"`
		Recommendation service `yaml:"recommendation"`
		Chat           service `yaml:"chat"`
	}

	service struct {
		URL    string `yaml:"url"`
		APIKey string `yaml:"api_key"`
	}

	swagger struct {
		Title       string `yaml:"title"`
		Description string `yaml:"description"`
		Version     string `yaml:"version"`
	}
)

func InitConfig() {
	configName := "config.dev"

	if os.Getenv("APP_ENV") != "" {
		configName += fmt.Sprintf(".%s", os.Getenv("APP_ENV"))
	}

	viper.SetDefault("pilot.mode", true)

	if err := viper.BindEnv("pilot.mode", "PILOT_MODE"); err != nil {
		log.Fatal(err)
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
		Services: services{
			Auth: service{
				URL:    viper.GetString("services.auth.url"),
				APIKey: viper.GetString("services.auth.api_key"),
			},
			Notification: service{
				URL:    viper.GetString("services.notification.url"),
				APIKey: viper.GetString("services.notification.api_key"),
			},
			Profile: service{
				URL:    viper.GetString("services.profile.url"),
				APIKey: viper.GetString("services.profile.api_key"),
			},
			Match: service{
				URL:    viper.GetString("services.match.url"),
				APIKey: viper.GetString("services.match.api_key"),
			},
			Recommendation: service{
				URL:    viper.GetString("services.recommendation.url"),
				APIKey: viper.GetString("services.recommendation.api_key"),
			},
			Chat: service{
				URL:    viper.GetString("services.chat.url"),
				APIKey: viper.GetString("services.chat.api_key"),
			},
		},
		Pilot: pilot{
			Mode: viper.GetBool("pilot.mode"),
		},
		Swagger: swagger{
			Title:       viper.GetString("swagger.title"),
			Description: viper.GetString("swagger.description"),
			Version:     viper.GetString("swagger.version"),
		},
	}
}
