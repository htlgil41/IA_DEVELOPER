package libs

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	Cloudflared CloudflaredConfig `mapstructure:"cloudflared"`
}

type CloudflaredConfig struct {
	Account string `mapstructure:"account"`
	APIKey  string `mapstructure:"apikey"`
}

func LoadConfigVyper() *Config {

	viper.SetConfigName("keys_IA_client")
	viper.SetConfigType("yml")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error al leer el archivo de configuración: %v", err)
		return nil
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalf("Error al decodificar la configuración: %v", err)
		return nil
	}

	return &config
}
