package main

import (
	"github.com/BurntSushi/toml"
	"io/ioutil"
)

type Config struct {
	BrokerUri string `toml:"broker.uri"`
	MetaTopic string `toml:"meta.topic"`
}

func LoadConfig(filename string) *Config {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	config := Config{}
	_, err = toml.Decode(data, &config)
	if err != nil {
		panic(err)
	}

	return &config
}
