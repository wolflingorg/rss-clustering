// Structure of config.ini file
package main

import (
	"gopkg.in/ini.v1"
	"log"
	"time"
)

type Config struct {
	LogPath    string
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
		log.Fatalf("Couldnt parse config file: %s\n", err)
	}
}
