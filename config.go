// Structure of config.ini file
package main

import (
	"gopkg.in/ini.v1"
	"time"
)

type Config struct {
	Clustering struct {
		Porog float64
	}
	Db struct {
		Host []string
	}
	Handler struct {
		Workers  int
		Tasks    int
		Interval time.Duration
	}
}

// Load and Map config from file
func LoadConfig(config *Config, CONFIG_PATH string) {
	err := ini.MapTo(config, CONFIG_PATH)
	if err != nil {
		panic("Could parse config file")
	}
}
