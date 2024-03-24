package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/agency"
	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/common"
)

// InitConfig Function that uses viper library to parse configuration
// parameters.  Viper is configured to read variables from both
// environment variables and the config file
// ./config.yaml. Environment variables takes precedence over
// parameters defined in the configuration file. If some of the
// variables cannot be parsed, an error is returned
func InitConfig() (*viper.Viper, error) {
	v := viper.New()

	// Configure viper to read env variables with the CLI_ prefix
	v.AutomaticEnv()
	v.SetEnvPrefix("cli")
	// Use a replacer to replace env variables underscores with
	// points. This let us use nested configurations in the config
	// file and at the same time define env variables for the
	// nested configurations
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Add env variables supported
	v.BindEnv("id")
	v.BindEnv("server", "address")
	v.BindEnv("loop", "period")
	v.BindEnv("loop", "lapse")
	v.BindEnv("log", "level")

	// The path to the file from which the bets will be read
	v.BindEnv("bets", "path")
	// The size of the batches/chunks
	v.BindEnv("batch", "size")

	// Try to read configuration from config file. If config file
	// does not exists then ReadInConfig will fail but
	// configuration can be loaded from the environment variables
	// so we shouldn't return an error in that case
	v.SetConfigFile("./config.yaml")
	if err := v.ReadInConfig(); err != nil {
		fmt.Printf("Configuration could not be read from config file. Using env variables instead")
	}

	// Parse time.Duration variables and return an error if those
	// variables cannot be parsed
	if _, err := time.ParseDuration(v.GetString("loop.lapse")); err != nil {
		return nil, errors.Wrapf(err, "Could not parse CLI_LOOP_LAPSE env var as time.Duration.")
	}

	if _, err := time.ParseDuration(v.GetString("loop.period")); err != nil {
		return nil, errors.Wrapf(err, "Could not parse CLI_LOOP_PERIOD env var as time.Duration.")
	}

	return v, nil
}

// InitLogger Receives the log level to be set in logrus as a
// string. This method parses the string and set the level to the
// logger. If the level string is not valid an error is returned
func InitLogger(logLevel string) error {
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		return err
	}

	customFormatter := &logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   false,
	}
	logrus.SetFormatter(customFormatter)
	logrus.SetLevel(level)
	return nil
}

// PrintConfig Print all the configuration parameters of the program.
// For debugging purposes only
func PrintConfig(v *viper.Viper) {
	logrus.Infof("action: config | result: success | client_id: %s | server_address: %s | loop_lapse: %v | loop_period: %v | log_level: %s",
		v.GetString("id"),
		v.GetString("server.address"),
		v.GetDuration("loop.lapse"),
		v.GetDuration("loop.period"),
		v.GetString("log.level"),
	)
}

type BufferedReadCloser struct {
	reader *bufio.Reader
	closer io.Closer
}

func (b *BufferedReadCloser) Read(p []byte) (int, error) {
	return b.reader.Read(p)
}

func (b *BufferedReadCloser) Close() error {
	return b.closer.Close()
}

func main() {
	v, err := InitConfig()
	if err != nil {
		log.Fatalf("%s", err)
	}

	if err := InitLogger(v.GetString("log.level")); err != nil {
		log.Fatalf("%s", err)
	}

	// Print program config with debugging purposes
	PrintConfig(v)

	clientConfig := common.ClientConfig{
		ServerAddress: v.GetString("server.address"),
		ID:            v.GetUint32("id"),
		BatchSize:     v.GetInt("batch.size"),
		LoopLapse:     v.GetDuration("loop.lapse"),
	}

	path := v.GetString("bets.path")
	f, err := os.Open(path)
	if err != nil {
		log.Fatalf("couldn't open bets file: %s", err.Error())
	}

	rc := &BufferedReadCloser{
		reader: bufio.NewReader(f),
		closer: f,
	}
	client := common.NewClient(rc, clientConfig)

	signalChannel := make(chan os.Signal, 1)
	doneChannel := make(chan struct{}, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
	client.HandleSignals(signalChannel, doneChannel)

	client.StartClientLoop()
}
