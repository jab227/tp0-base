package dockergen_test

import (
	"dockergen"
	"fmt"
	"testing"
)
// TODO(juan): Parametrize image, entrypoint and enviroment tags
func TestNewDockerCompose(t *testing.T) {
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
        - subnet: 172.25.125.0/24`
		cfg := dockergen.DockerComposeConfig{
			Name:    "tp0",
			Version: "3.9",
		}
		got := dockergen.GenerateDockerCompose(cfg)
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
      - CLI_ID=1
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
      - CLI_ID=1
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
        - subnet: 172.25.125.0/24`
		cfg := dockergen.DockerComposeConfig{
			Name:    "tp0",
			Version: "3.9",
			Clients: 3,
		}
		got := dockergen.GenerateDockerCompose(cfg)
		fmt.Println(got)
		if got != want {
			t.Fatalf("got %s, want %s", got, want)
		}
	})
}
