package main

import (
	"bytes"
	"testing"
)

func TestNewConfigForEnv(t *testing.T) {
	t.Run("env vars are set correctly", func(t *testing.T) {
		lookupFunc := func(s string) (string, bool) {
			switch s {
			case EnvVarFilename:
				return "test.yaml", true
			case EnvVarComposeName:
				return "tp1", true
			case EnvVarNumberOfClients:
				return "10", true
			}
			return "", false
		}

		want := DockerComposeConfig{
			Name:     "tp1",
			Filename: "test.yaml",
			Clients:  10,
		}
		got, err := NewConfigFromEnv(lookupFunc)
		if err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}

		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("env vars are not defined", func(t *testing.T) {
		lookupFunc := func(s string) (string, bool) {
			return "", false
		}

		want := DockerComposeConfig{
			Name:     ComposeNameDefault,
			Filename: FilenameDefault,
			Clients:  ClientsDefault,
		}
		got, err := NewConfigFromEnv(lookupFunc)
		if err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}

		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("empty filename", func(t *testing.T) {
		lookupFunc := func(s string) (string, bool) {
			switch s {
			case EnvVarFilename:
				return "", true
			case EnvVarComposeName:
				return "tp1", true
			case EnvVarNumberOfClients:
				return "10", true
			}
			return "", false
		}

		_, err := NewConfigFromEnv(lookupFunc)
		if err == nil {
			t.Error("Expected error")
		}
	})

	t.Run("empty compose name", func(t *testing.T) {
		lookupFunc := func(s string) (string, bool) {
			switch s {
			case EnvVarFilename:
				return "test.yaml", true
			case EnvVarComposeName:
				return "", true
			case EnvVarNumberOfClients:
				return "10", true
			}
			return "", true
		}

		_, err := NewConfigFromEnv(lookupFunc)
		if err == nil {
			t.Error("Expected error")
		}
	})

	t.Run("invalid number of clients", func(t *testing.T) {
		lookupFunc := func(s string) (string, bool) {
			switch s {
			case EnvVarFilename:
				return "test.yaml", true
			case EnvVarComposeName:
				return "", true
			case EnvVarNumberOfClients:
				return "abc", true
			}
			return "", true
		}

		_, err := NewConfigFromEnv(lookupFunc)
		if err == nil {
			t.Error("Expected error")
		}
	})
}

func TestGenerateDockerCompose(t *testing.T) {
	t.Run("zero clients specified", func(t *testing.T) {
		const want = `version: '3.9'
name: tp0
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
		cfg := DockerComposeConfig{
			Name: "tp0",
		}
		var buf bytes.Buffer
		GenerateDockerCompose(cfg, &buf)
		got := buf.String()
		if got != want {
			t.Fatalf("got %s, want %s", got, want)
		}
	})

	t.Run("one client specified", func(t *testing.T) {
		const want = `version: '3.9'
name: tp0
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
		cfg := DockerComposeConfig{
			Name:    "tp0",
			Clients: 3,
		}
		var buf bytes.Buffer
		GenerateDockerCompose(cfg, &buf)
		got := buf.String()
		if got != want {
			t.Fatalf("got %s, want %s", got, want)
		}
	})
}
