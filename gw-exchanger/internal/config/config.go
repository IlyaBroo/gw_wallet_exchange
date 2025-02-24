package config

import (
	"io/ioutil"

	"gw-exchanger/internal/logger"

	yaml "gopkg.in/yaml.v2"
)

type ConfigAdr struct {
	Database_url string `yaml:"database_url"`
	APP_ADR      string `yaml:"app_adr"`
	Cache_ttl    int    `yaml:"cache_ttl"`
}

func LoadConfig(filePath string) (*logger.Config, *ConfigAdr, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, nil, err
	}
	cfg := new(logger.Config)
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, nil, err
	}
	cfgAdr := new(ConfigAdr)
	if err := yaml.Unmarshal(data, cfgAdr); err != nil {
		return nil, nil, err
	}
	return cfg, cfgAdr, nil
}
