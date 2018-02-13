package main

import (
	"github.com/BurntSushi/toml"
	"io/ioutil"
)

type BrokerConfig struct {
	Uri string `toml:"uri"`
}

type MetaConfig struct {
	Topic string `toml:"topic"`
}

type Config struct {
	Broker BrokerConfig `toml:"broker"`
	Meta   MetaConfig   `toml:"meta"`
}

func LoadConfig(filename string) *Config {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	config := Config{}
	_, err = toml.Decode(string(data), &config)
	if err != nil {
		panic(err)
	}

	return &config
}
