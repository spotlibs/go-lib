//go:build ignore

package queue

import (
	"fmt"
	"sync"

	"github.com/RichardKnop/machinery/v2"
	redisbackend "github.com/RichardKnop/machinery/v2/backends/redis"
	redisbroker "github.com/RichardKnop/machinery/v2/brokers/redis"
	"github.com/RichardKnop/machinery/v2/config"
	"github.com/RichardKnop/machinery/v2/locks/eager"
	"github.com/RichardKnop/machinery/v2/log"

	logcontract "github.com/goravel/framework/contracts/log"
	"github.com/goravel/framework/support/color"
)

var (
	machineryServer     *machinery.Server
	machineryServerOnce sync.Once
)

type Machinery struct {
	config *Config
	log    logcontract.Log
}

func NewMachinery(config *Config, log logcontract.Log) *Machinery {
	return &Machinery{config: config, log: log}
}

func (m *Machinery) Server(connection string, queue string) (*machinery.Server, error) {
	driver := m.config.Driver(connection)

	switch driver {
	case DriverSync:
		color.Yellow().Println("Queue sync driver doesn't need to be run")

		return nil, nil
	case DriverRedis:
		return m.redisServer(connection, queue), nil
	}

	return nil, fmt.Errorf("unknown queue driver: %s", driver)
}

func (m *Machinery) redisServer(connection string, queue string) *machinery.Server {
	machineryServerOnce.Do(func() {
		redisConfig, database, defaultQueue := m.config.Redis(connection)
		if queue == "" {
			queue = defaultQueue
		}

		cnf := &config.Config{
			DefaultQueue: queue,
			Redis:        &config.RedisConfig{},
		}

		debug := m.config.config.GetBool("app.debug")
		log.DEBUG = NewDebug(debug, m.log)
		log.INFO = NewInfo(debug, m.log)
		log.WARNING = NewWarning(debug, m.log)
		log.ERROR = NewError(debug, m.log)
		log.FATAL = NewFatal(debug, m.log)

		broker := redisbroker.NewGR(cnf, []string{redisConfig}, database)
		backend := redisbackend.NewGR(cnf, []string{redisConfig}, database)
		machineryServer = machinery.NewServer(cnf, broker, backend, eager.New())
	})
	return machineryServer
}
