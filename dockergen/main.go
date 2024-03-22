package main

import (
	"log"
	"os"
)

func main() {
	cfg, err := NewConfigFromEnv(os.LookupEnv)
	if err != nil {
		log.Fatalf("couldn't create config from env: %s", err.Error())
	}
	f, err := os.OpenFile(cfg.Filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatalf("couldn't create/open file: %s", err.Error())
	}
	defer f.Close()
	GenerateDockerCompose(cfg, f)
}
