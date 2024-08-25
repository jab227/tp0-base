package main

import (
	"log"
	"os"
)

func main() {
	var cfg Config
	InitConfig(&cfg)
	f, err := os.OpenFile(cfg.Filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatalf("couldn't create/open file: %s", err.Error())
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Printf("couldn't close file correctly: %s", err)
		}
	}()

	GenerateCompose(cfg, f)
}
