package configs

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

var Config *config

type (
	config struct {
		App      app      `yaml:"app"`
		Services services `yaml:"services"`
		Swagger  swagger  `yaml:"swagger"`
	}

	app struct {
		Name        string `yaml:"name"`
		Environment string `yaml:"environment"`
		Url         string `yaml:"url"`
		Key         string `yaml:"key"`
	}

	services struct {
		Auth         service `yaml:"auth"`
		Notification service `yaml:"notification"`
		Profile      service `yaml:"profile"`
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
	viper.SetConfigType("yaml")
	
	configNames := []string{"config.dev", "config.dev.development", "config"}
	
	viper.AddConfigPath(".")
	viper.AddConfigPath("../../..")
	viper.AddConfigPath("/app")
	
	var err error
	found := false
	
	for _, name := range configNames {
		viper.SetConfigName(name)
		if err = viper.ReadInConfig(); err == nil {
			found = true
			fmt.Printf("Config file loaded: %s\n", name)
			break
		}
	}
	
	if !found {
		fmt.Printf("Warning: Config file not found. Using defaults. Error: %s\n", err)
		setDefaults()
	}

	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	Config = &config{
		App: app{
			Name:        getString("app.name", "api-gateway"),
			Environment: getString("app.environment", env),
			Url:         getString("app.url", "0.0.0.0:3000"),
			Key:         getString("app.key", ""),
		},
		Services: services{
			Auth: service{
				URL:    getString("services.auth.url", "http://auth-winx:3001"),
				APIKey: getString("services.auth.api_key", ""),
			},
			Notification: service{
				URL:    getString("services.notification.url", "http://notification-winx:3002"),
				APIKey: getString("services.notification.api_key", ""),
			},
			Profile: service{
				URL:    getString("services.profile.url", "http://profile-winx:3003"),
				APIKey: getString("services.profile.api_key", ""),
			},
		},
		Swagger: swagger{
			Title:       getString("swagger.title", "Winx API Gateway"),
			Description: getString("swagger.description", "API gateway entrypoint for Winx services."),
			Version:     getString("swagger.version", "1.0.0"),
		},
	}
}

func setDefaults() {
	viper.SetDefault("app.name", "api-gateway")
	viper.SetDefault("app.environment", "development")
	viper.SetDefault("app.url", "0.0.0.0:3000")
	viper.SetDefault("app.key", "")
	viper.SetDefault("services.auth.url", "http://auth-winx:3001")
	viper.SetDefault("services.auth.api_key", "")
	viper.SetDefault("services.notification.url", "http://notification-winx:3002")
	viper.SetDefault("services.notification.api_key", "")
	viper.SetDefault("services.profile.url", "http://profile-winx:3003")
	viper.SetDefault("services.profile.api_key", "")
	viper.SetDefault("swagger.title", "Winx API Gateway")
	viper.SetDefault("swagger.description", "API gateway entrypoint for Winx services.")
	viper.SetDefault("swagger.version", "1.0.0")
}

func getString(key, defaultValue string) string {
	if viper.IsSet(key) {
		return viper.GetString(key)
	}
	return defaultValue
}