package dockergen

import (
	"fmt"
	"strings"
)

type DockerComposeConfig struct {
	Name    string
	Version string
	Clients int
}

const serverService = `server:
    container_name: server
    image: server:latest
    entrypoint: python3 /main.py
    environment:
      - PYTHONUNBUFFERED=1
      - LOGGING_LEVEL=DEBUG
    networks:
      - testing_net`

const networksString = `networks:
  testing_net:
    ipam:
      driver: default
      config:
        - subnet: 172.25.125.0/24`

const sharedTagsClients = `    image: client:latest
    entrypoint: /client
    environment:
      - CLI_ID=1
      - CLI_LOG_LEVEL=DEBUG
    networks:
      - testing_net
    depends_on:
      - server`

func GenerateDockerCompose(cfg DockerComposeConfig) string {
	var builder strings.Builder
	WriteVersion(&builder, cfg.Version)
	WriteLabel(&builder, cfg.Name)
	builder.WriteString("services:\n")
	serverServiceString := fmt.Sprintf("  %s\n", serverService)
	builder.WriteString(serverServiceString)
	builder.WriteByte('\n')			
	for i := 0; i < cfg.Clients; i++ {
		clientName := fmt.Sprintf("client%d", i + 1)
		builder.WriteString("  ")
		builder.WriteString(clientName)
		builder.WriteByte(':')
		builder.WriteByte('\n')		
		containerName := fmt.Sprintf("    container_name: %s\n", clientName)
		builder.WriteString(containerName)		
		builder.WriteString(sharedTagsClients)
		builder.WriteByte('\n')
		builder.WriteByte('\n')						
	}
	builder.WriteString(networksString)
	return builder.String()
}

func WriteLabel(builder *strings.Builder, label string) {
	builder.WriteString("name: ")
	builder.WriteString(label)
	builder.WriteByte('\n')
}

func WriteVersion(builder *strings.Builder, version string) {
	builder.WriteString("version: ")
	builder.WriteByte('\'')
	builder.WriteString(version)
	builder.WriteByte('\'')
	builder.WriteByte('\n')
}

func main() {
	fmt.Println("Hello world!")
}
