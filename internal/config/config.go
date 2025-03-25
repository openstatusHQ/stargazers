package config

import (
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

var k = koanf.New(".")

type Config struct {
	Repositories []struct {
		Name  string `koanf:"name"`
		Owner string `koanf:"owner"`
	} `koanf:"repositories"`
}

func ReadConfig(path string) (*Config, error) {

	file := file.Provider(path)

	err := k.Load(file, yaml.Parser())

	if err != nil {
		return nil, err
	}

	var out Config

	err = k.Unmarshal("", &out)
	if err != nil {
		return nil, err
	}

	return &out, nil
}
