package configs

import (
	"github.com/spf13/viper"
)

type conf struct {
	Environment          string `mapstructure:"ENVIRONMENT"`
	WebServerPort        string `mapstructure:"WEB_SERVER_PORT"`
	WeatherApiKey        string `mapstructure:"WEATHER_API_KEY"`
	ExternalCallURL      string `mapstructure:"EXTERNAL_CALL_URL"`
	OtelServiceName      string `mapstructure:"OTEL_SERVICE_NAME"`
	OtelExporterEndpoint string `mapstructure:"OTEL_EXPORTER_OTLP_ENDPOINT"`
}

func LoadConfig() (*conf, error) {
	viper.AutomaticEnv()
	viper.BindEnv("ENVIRONMENT")
	viper.BindEnv("WEB_SERVER_PORT")
	viper.BindEnv("WEATHER_API_KEY")
	viper.BindEnv("EXTERNAL_CALL_URL")           // URL do microsservico clima (service discovery do compose)
	viper.BindEnv("OTEL_SERVICE_NAME")           // nome do servico exibido no Zipkin
	viper.BindEnv("OTEL_EXPORTER_OTLP_ENDPOINT") // endpoint gRPC do OTEL Collector
	viper.BindEnv("PORT")                        // Cloud Run injeta PORT

	isProduction := viper.GetString("ENVIRONMENT") == "production"

	// Em production (Cloud Run) nao ha .env: a config vem de env vars/Secret Manager.
	// Fora de production o .env e obrigatorio -- falha explicita para pegar misconfig.
	viper.SetConfigFile(".env")
	if err := viper.ReadInConfig(); err != nil && !isProduction {
		return nil, err
	}

	var cfg conf
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	// Cloud Run manda apenas PORT; usa como fallback do WEB_SERVER_PORT.
	if cfg.WebServerPort == "" {
		cfg.WebServerPort = viper.GetString("PORT")
	}
	return &cfg, nil
}
