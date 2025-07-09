package config

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/viper"
	"k8s.io/client-go/util/homedir"
)

type Config struct {
	Kubernetes KubernetesConfig `mapstructure:"kubernetes"`
	ArgoCD     ArgoCDConfig     `mapstructure:"argocd"`
}

type KubernetesConfig struct {
	ConfigPath string `mapstructure:"config_path"`
	Namespace  string `mapstructure:"namespace"`
}

type ArgoCDConfig struct {
	Server    string `mapstructure:"server"`
	Token     string `mapstructure:"token"`
	Namespace string `mapstructure:"namespace"`
}

func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.k8s-argocd-manager")

	// Set defaults
	viper.SetDefault("kubernetes.namespace", "default")
	viper.SetDefault("argocd.server", "localhost:8080")
	viper.SetDefault("argocd.namespace", "argocd")

	if home := homedir.HomeDir(); home != "" {
		viper.SetDefault("kubernetes.config_path", filepath.Join(home, ".kube", "config"))
	}

	// Read environment variables
	viper.SetEnvPrefix("K8S_ARGOCD")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}
