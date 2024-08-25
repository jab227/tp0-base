package main

import (
	"bytes"
	"testing"
)

func TestGenerateDockerCompose(t *testing.T) {
	t.Run("zero clients specified", func(t *testing.T) {
		const want = `name: tp0
services:
  server:
    container_name: server
    image: server:latest
    entrypoint: python3 /main.py
    environment:
      - PYTHONUNBUFFERED=1
      - LOGGING_LEVEL=DEBUG
    networks:
      - testing_net
networks:
  testing_net:
    ipam:
      driver: default
      config:
        - subnet: 172.25.125.0/24
`
		cfg := Config{Clients: 0}
		var buf bytes.Buffer
		GenerateCompose(cfg, &buf)
		got := buf.String()
		if got != want {
			t.Fatalf("got %s, want %s", got, want)
		}
	})

	t.Run("more than one client specified", func(t *testing.T) {
		const want = `name: tp0
services:
  server:
    container_name: server
    image: server:latest
    entrypoint: python3 /main.py
    environment:
      - PYTHONUNBUFFERED=1
      - LOGGING_LEVEL=DEBUG
    networks:
      - testing_net
  client1:
    container_name: client1
    image: client:latest
    entrypoint: /client
    environment:
      - CLI_ID=1
      - CLI_LOG_LEVEL=DEBUG
    networks:
      - testing_net
    depends_on:
      - server
  client2:
    container_name: client2
    image: client:latest
    entrypoint: /client
    environment:
      - CLI_ID=2
      - CLI_LOG_LEVEL=DEBUG
    networks:
      - testing_net
    depends_on:
      - server
  client3:
    container_name: client3
    image: client:latest
    entrypoint: /client
    environment:
      - CLI_ID=3
      - CLI_LOG_LEVEL=DEBUG
    networks:
      - testing_net
    depends_on:
      - server
networks:
  testing_net:
    ipam:
      driver: default
      config:
        - subnet: 172.25.125.0/24
`
		cfg := Config{
			Clients: 3,
		}
		var buf bytes.Buffer
		GenerateCompose(cfg, &buf)
		got := buf.String()
		if got != want {
			t.Fatalf("got %s, want %s", got, want)
		}
	})
}


