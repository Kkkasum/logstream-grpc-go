package config

import (
	"log"
	"path/filepath"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/file"
)

type Config struct {
	ServerConfig *ServerConfig `json:"server"`
	DBConfig     *DBConfig     `json:"db"`
}

type ServerConfig struct {
	Host string `json:"Host"`
	Port int    `json:"port"`
}

type DBConfig struct {
	Host     string `json:"host"`
	User     string `json:"user"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Port     int    `json:"port"`
}

func Load(configPath string) (*Config, error) {
	k := koanf.New(".")

	err := k.Load(confmap.Provider(defaultConfig, "."), nil)
	if err != nil {
		log.Printf("failed to load default config; err: %v", err)
		return nil, err
	}

	if configPath != "" {
		path, err := filepath.Abs(configPath)
		if err != nil {
			log.Printf("failed to get absolute config path; configPath: %s, err: %v", configPath, err)
			return nil, err
		}
		log.Printf("Load config file from %s", path)
		if err := k.Load(file.Provider(path), yaml.Parser()); err != nil {
			log.Printf("failed to load config from file; err: %v", err)
			return nil, err
		}
	}

	var cfg Config
	if err := k.UnmarshalWithConf("", &cfg, koanf.UnmarshalConf{Tag: "json", FlatPaths: false}); err != nil {
		log.Printf("failed to unmarshal with conf; err: %v", err)
		return nil, err
	}

	return &cfg, err
}
