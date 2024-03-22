package main

import (
	"fmt"
	"io"
	"strconv"
	"strings"
)

type DockerComposeConfig struct {
	Filename string
	Name     string
	Clients  int
}

func GenerateDockerCompose(cfg DockerComposeConfig, w io.Writer) {
	var builder strings.Builder
	writeVersion(&builder, "3.9")
	writeKeyValue(&builder, 0, pair{key: "name", value: cfg.Name})
	writeKeyValue(&builder, 0, pair{key: "services"})
	writeServer(&builder)
	for i := 0; i < cfg.Clients; i++ {
		id := i + 1
		writeClient(&builder, id)
	}
	writeNetworkConfig(&builder)
	w.Write([]byte(builder.String()))
}

func writeNetworkConfig(builder *strings.Builder) {
	writeKeyValue(builder, 0, pair{key: "networks"})
	writeKeyValue(builder, 2, pair{key: "testing_net"})
	writeKeyValue(builder, 4, pair{key: "ipam"})
	writeKeyValue(builder, 6, pair{key: "driver", value: "default"})
	writeKeyValue(builder, 6, pair{key: "config"})
	writeItemList(builder, 8, "subnet: 172.25.125.0/24")
}

func writeClient(builder *strings.Builder, id int) {
	clientName := fmt.Sprintf("client%d", id)
	clientIdEnv := fmt.Sprintf("CLI_ID=%d", id)

	writeKeyValue(builder, 2, pair{key: clientName})
	writeKeyValue(builder, 4, pair{key: "container_name", value: clientName})
	writeKeyValue(builder, 4, pair{key: "image", value: "client:latest"})
	writeKeyValue(builder, 4, pair{key: "entrypoint", value: "/client"})
	writeKeyValue(builder, 4, pair{key: "environment"})
	writeItemList(builder, 6, clientIdEnv, "CLI_LOG_LEVEL=DEBUG")
	writeKeyValue(builder, 4, pair{key: "networks"})
	writeItemList(builder, 6, "testing_net")
	writeKeyValue(builder, 4, pair{key: "depends_on"})
	writeItemList(builder, 6, "server")
}

type pair struct {
	key, value string
}

func writeItemList(builder *strings.Builder, ident int, values ...string) {
	spaces := strings.Repeat(" ", ident)
	for _, v := range values {
		builder.WriteString(spaces)
		builder.WriteByte('-')
		builder.WriteByte(' ')
		builder.WriteString(v)
		builder.WriteByte('\n')
	}
}

func writeKeyValue(builder *strings.Builder, ident int, p pair) {
	spaces := strings.Repeat(" ", ident)
	builder.WriteString(spaces)
	builder.WriteString(p.key)
	builder.WriteByte(':')
	if len(p.value) != 0 {
		builder.WriteByte(' ')
		builder.WriteString(p.value)
	}
	builder.WriteByte('\n')
}

func writeServer(builder *strings.Builder) {
	writeKeyValue(builder, 2, pair{key: "server"})
	writeKeyValue(builder, 4, pair{key: "container_name", value: "server"})
	writeKeyValue(builder, 4, pair{key: "image", value: "server:latest"})
	writeKeyValue(builder, 4, pair{key: "entrypoint", value: "python3 /main.py"})
	writeKeyValue(builder, 4, pair{key: "environment"})
	writeItemList(builder, 6, "PYTHONUNBUFFERED=1", "LOGGING_LEVEL=DEBUG")
	writeKeyValue(builder, 4, pair{key: "networks"})
	writeItemList(builder, 6, "testing_net")
}

func writeVersion(builder *strings.Builder, version string) {
	builder.WriteString("version: ")
	builder.WriteByte('\'')
	builder.WriteString(version)
	builder.WriteByte('\'')
	builder.WriteByte('\n')
}

type LookupEnvFunc func(string) (string, bool)

const (
	EnvVarNumberOfClients = "DOCKERGEN_NUMBER_OF_CLIENTS"
	EnvVarFilename        = "DOCKERGEN_FILENAME"
	EnvVarComposeName     = "DOCKERGEN_COMPOSE_NAME"
)

const (
	ClientsDefault     = 1
	FilenameDefault    = "docker-compose-dev.yaml"
	ComposeNameDefault = "tp0"
)

func NewConfigFromEnv(lookup LookupEnvFunc) (DockerComposeConfig, error) {
	numberOfClients := ClientsDefault
	value, ok := lookup(EnvVarNumberOfClients)
	if ok {
		n, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			err = fmt.Errorf("%s: %w", EnvVarNumberOfClients, err)
			return DockerComposeConfig{}, err
		}
		numberOfClients = int(n)
	}

	value, ok = lookup(EnvVarFilename)
	filename := FilenameDefault
	if ok {
		if len(value) == 0 {
			err := fmt.Errorf("%s: expected non-empty filename", EnvVarFilename)
			return DockerComposeConfig{}, err
		}
		filename = value
	}

	value, ok = lookup(EnvVarComposeName)
	name := ComposeNameDefault
	if ok {
		if len(value) == 0 {
			err := fmt.Errorf("%s: expected non-empty compose name", EnvVarComposeName)
			return DockerComposeConfig{}, err
		}
		name = value
	}

	return DockerComposeConfig{
		Filename: filename,
		Clients:  numberOfClients,
		Name:     name,
	}, nil
}
