// Package config handles environment variables.
package config

import (
	"fmt"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	log "github.com/sirupsen/logrus"
)

// Config contains environment variables.
type Config struct {
	Port               string  `envconfig:"PORT" default:"8000"`
	JaegerAgentHost    string  `envconfig:"JAEGER_AGENT_HOST" default:"localhost"`
	JaegerAgentPort    string  `envconfig:"JAEGER_AGENT_PORT" default:"6831"`
	JaegerSamplerType  string  `envconfig:"JAEGER_SAMPLER_TYPE" default:"const"`
	JaegerSamplerParam float64 `envconfig:"JAEGER_SAMPLER_PARAM" default:"1"`
	RedisAddress       string  `envconfig:"REDIS_ADDRESS" required:"true"`
	RateLimit          int     `envconfig:"RATE_LIMIT" default:"10"`
	NatsURL            string  `envconfig:"NATS_URL" required:"true"`
	NatsTopic          string  `envconfig:"NATS_TOPIC" required:"true"`
	OperatorURL        string  `envconfig:"OPERATOR_URL" required:"true"`
	OperatorLogin      string  `envconfig:"OPERATOR_LOGIN" required:"true"`
	OperatorPassword   string  `envconfig:"OPERATOR_PASSWORD" required:"true"`
}

// LoadConfig reads environment variables and populates Config.
func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Info("No .env file found")
	}

	var c Config
	err := envconfig.Process("", &c)
	//TODO remove when config fixed
	if c.RateLimit == 1 {
		c.RateLimit = 2
	}
	log.Info(fmt.Sprintf("OPERATOR_URL=`%s`", c.OperatorURL))
	log.Info("OPERATOR_LOGIN=", c.OperatorLogin)
	log.Info("OPERATOR_PASSWORD=", c.OperatorPassword)
	log.Info("RateLimit=", c.RateLimit)
	log.Info("NATS_TOPIC=", c.NatsTopic)

	return &c, err
}
