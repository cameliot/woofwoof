package main

import (
	"github.com/BurntSushi/toml"
	"github.com/cameliot/alpaca"

	"io/ioutil"
)

type BrokerConfig struct {
	Uri string `toml:"uri"`
}

type MetaConfig struct {
	Topic string `toml:"topic"`
}

type WatchConfig struct {
	Group string `toml:"group"`

	Request string `toml:"request"`
	Success string `toml:"success"`
	Error   string `toml:"error"`
}

type ServiceConfig struct {
	Handle  string        `toml:"handle"`
	Topic   string        `toml:"topic"`
	Watches []WatchConfig `toml:"watch"`
}

type Config struct {
	Herd   string       `toml:"name"`
	Broker BrokerConfig `toml:"broker"`
	Meta   MetaConfig   `toml:"meta"`

	Services []ServiceConfig `toml:"services"`
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

/*
Create alpaca routes from config
*/
func (self *Config) AlpacaRoutes() alpaca.Routes {
	routes := alpaca.Routes{
		"meta": self.Meta.Topic,
	}

	for _, service := range self.Services {
		routes[service.Handle] = service.Topic
	}

	return routes
}
