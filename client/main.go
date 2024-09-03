package main

import (
	"fmt"
	"github.com/op/go-logging"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/agency"
	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/batch"
	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/common"
)

var log = logging.MustGetLogger("log")

// InitConfig Function that uses viper library to parse configuration parameters.
// Viper is configured to read variables from both environment variables and the
// config file ./config.yaml. Environment variables takes precedence over parameters
// defined in the configuration file. If some of the variables cannot be parsed,
// an error is returned
func InitConfig() (*viper.Viper, error) {
	v := viper.New()

	// Configure viper to read env variables with the CLI_ prefix
	v.AutomaticEnv()
	v.SetEnvPrefix("cli")
	// Use a replacer to replace env variables underscores with points. This let us
	// use nested configurations in the config file and at the same time define
	// env variables for the nested configurations
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Add env variables supported
	v.BindEnv("id")
	v.BindEnv("server", "address")
	v.BindEnv("loop", "period")
	v.BindEnv("loop", "amount")
	v.BindEnv("log", "level")

	// Bet env variables
	v.BindEnv("bettor", "nombre")
	v.BindEnv("bettor", "apellido")
	v.BindEnv("bettor", "documento")
	v.BindEnv("bettor", "nacimiento")
	v.BindEnv("bettor", "numbero")
	// Try to read configuration from config file. If config file
	// does not exists then ReadInConfig will fail but configuration
	// can be loaded from the environment variables so we shouldn't
	// return an error in that case
	v.SetConfigFile("./config.yaml")
	if err := v.ReadInConfig(); err != nil {
		fmt.Printf("Configuration could not be read from config file. Using env variables instead")
	}

	// Parse time.Duration variables and return an error if those variables cannot be parsed

	if _, err := time.ParseDuration(v.GetString("loop.period")); err != nil {
		return nil, errors.Wrapf(err, "Could not parse CLI_LOOP_PERIOD env var as time.Duration.")
	}

	return v, nil
}

// InitLogger Receives the log level to be set in go-logging as a string. This method
// parses the string and set the level to the logger. If the level string is not
// valid an error is returned
func InitLogger(logLevel string) error {
	baseBackend := logging.NewLogBackend(os.Stdout, "", 0)
	format := logging.MustStringFormatter(
		`%{time:2006-01-02 15:04:05} %{level:.5s}     %{message}`,
	)
	backendFormatter := logging.NewBackendFormatter(baseBackend, format)

	backendLeveled := logging.AddModuleLevel(backendFormatter)
	logLevelCode, err := logging.LogLevel(logLevel)
	if err != nil {
		return err
	}
	backendLeveled.SetLevel(logLevelCode, "")

	// Set the backends to be used.
	logging.SetBackend(backendLeveled)
	return nil
}

// PrintConfig Print all the configuration parameters of the program.
// For debugging purposes only
func PrintConfig(v *viper.Viper) {
	log.Infof("action: config | result: success | client_id: %s | server_address: %s | loop_amount: %v | loop_period: %v | log_level: %s",
		v.GetString("id"),
		v.GetString("server.address"),
		v.GetInt("loop.amount"),
		v.GetDuration("loop.period"),
		v.GetString("log.level"),
	)
}

func PrintBettor(v *viper.Viper) {
	log.Infof("action: bet | result: success | client_id: %s | nombre: %s | apellido: %s | nacimiento: %s | documento: %s | numero: %s",
		v.GetString("id"),
		v.GetString("bettor.nombre"),
		v.GetString("bettor.apellido"),
		v.GetString("bettor.nacimiento"),
		v.GetString("bettor.documento"),
		v.GetString("bettor.numero"))
}

func NewBetFromEnv(v *viper.Viper) agency.Bettor {
	bettor := agency.Bettor{
		Name:      v.GetString("bettor.nombre"),
		Surname:   v.GetString("bettor.apellido"),
		DNI:       v.GetString("bettor.documento"),
		Birthdate: v.GetString("bettor.nacimiento"),
		BetNumber: v.GetString("bettor.numero"),
	}
	return bettor
}

func main() {
	// v, err := InitConfig()
	// if err != nil {
	// 	log.Criticalf("%s", err)
	// }

	// if err := InitLogger(v.GetString("log.level")); err != nil {
	// 	log.Criticalf("%s", err)
	// }

	// // Print program config with debugging purposes
	// PrintConfig(v)
	// PrintBettor(v)
	// clientConfig := common.ClientConfig{
	// 	ServerAddress: v.GetString("server.address"),
	// 	ID:            v.GetString("id"),
	// 	LoopAmount:    v.GetInt("loop.amount"),
	// 	LoopPeriod:    v.GetDuration("loop.period"),
	// }
	// bettor := NewBetFromEnv(v)
	// handler := common.NewSignalHandler()
	// client := common.NewClient(clientConfig, bettor, handler.Done())
	// go handler.Run()
	// client.StartClientLoop()
	f, _ := os.Open("a1.csv")

	handler := common.NewSignalHandler()
	batcher := batch.NewBatcher(10, 8*1024)
	go handler.Run()
	var wg sync.WaitGroup
	wg.Add(1)
	bets := common.BetScannerRun(f, &wg, handler.Done())
	wg.Add(1)
	chunks := common.BatchProcessor(batcher, &wg, bets, handler.Done())
	i := 0
	for {
		select {
		case <-handler.Done():
			wg.Wait()
			return
		case r, more := <-chunks:
			if r.Err != nil {
				fmt.Println(r.Err)
				return
			}
			if !more {
				fmt.Println("exit")
				return
			}
			i += 1
			fmt.Println(i)
		}
	}
}
