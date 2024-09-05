package main

import (
	"flag"
	"fmt"
	"io"
	"strings"
)

const (
	ClientsDefault     = 1
	FilenameDefault    = "docker-compose-dev.yaml"
	ComposeNameDefault = "tp0"
)

type pair struct {
	key, value string
}

type Config struct {
	Filename string
	Clients  int
}

func GenerateCompose(cfg Config, w io.Writer) {
	var builder strings.Builder
	writeKeyValue(&builder, 0, pair{key: "name", value: "tp0"})
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

	writeKeyValue(builder, 2, pair{key: clientName})
	writeKeyValue(builder, 4, pair{key: "container_name", value: clientName})
	writeKeyValue(builder, 4, pair{key: "image", value: "client:latest"})
	writeKeyValue(builder, 4, pair{key: "entrypoint", value: "/client"})
	writeKeyValue(builder, 4, pair{key: "networks"})
	writeItemList(builder, 6, "testing_net")
	writeKeyValue(builder, 4, pair{key: "depends_on"})
	writeItemList(builder, 6, "server")
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
	writeKeyValue(builder, 4, pair{key: "entrypoint", value: "python3 ./main.py"})
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

func InitConfig(cfg *Config) {
	flag.IntVar(&cfg.Clients, "clients", ClientsDefault, "number of clients")
	flag.StringVar(&cfg.Filename, "filename", FilenameDefault, "name of the output file")
	flag.Parse()
}
