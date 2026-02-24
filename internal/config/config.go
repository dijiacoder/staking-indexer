package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Database Database `mapstructure:"database"`
	Ethereum Ethereum `mapstructure:"ethereum"`
	Scanner  Scanner  `mapstructure:"scanner"`
}

type Database struct {
	DSN   string `mapstructure:"dsn"`
	Debug bool   `mapstructure:"debug"`
}

type Ethereum struct {
	RPCURL        string `mapstructure:"rpc_url"`
	ChainID       int64  `mapstructure:"chain_id"`
	ContractAddr  string `mapstructure:"contract_address"`
	Confirmations int64  `mapstructure:"confirmations"`
}

type Scanner struct {
	BatchSize    int `mapstructure:"batch_size"`
	ScanInterval int `mapstructure:"scan_interval"`
	ScanTimeout  int `mapstructure:"scan_timeout"`
}

func LoadConfig(configPath string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(configPath)
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}
