package main

import (
	"fmt"
	"log"

	"rtb-platform/pkg/config"
)

type AppConfig struct {
	Server    ServerConfig `yaml:"server"`
	Aerospike struct {
		Host string `yaml:"host" env:"AEROSPIKE_HOST"`
		Port int    `yaml:"port" env:"AEROSPIKE_PORT"`
	} `yaml:"aerospike"`
}

type ServerConfig struct {
	Port int `yaml:"port" env:"SERVER_PORT"`
}

func main() {
	var cfg AppConfig
	if err := config.Load(&cfg, config.WithPath("../../../configs/auction/dev.yaml")); err != nil {
		log.Fatalf("cannot load config: %v", err)
	}
	fmt.Printf("Loaded: %+v\n", cfg)
}
