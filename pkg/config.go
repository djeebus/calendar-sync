package pkg

import (
	"github.com/caarlos0/env/v11"
	"github.com/pkg/errors"
)

type Config struct {
	ClientSecretsPath string `env:"CS_CLIENT_SECRETS_PATH,required"`
	Listen            string `env:"CS_LISTEN" envDefault:":31450"`

	DatabaseDriver string `env:"CS_DATABASE_DRIVER" envDefault:"sqlite3"`
	DatabaseSource string `env:"CS_DATABASE_SOURCE" envDefault:"./database.db"`

	TemporalHostPort  string `env:"CS_TEMPORAL_HOSTPORT"`
	TemporalNamespace string `env:"CS_TEMPORAL_NAMESPACE"`
	TemporalIdentity  string `env:"CS_TEMPORAL_IDENTITY"`
	TemporalTaskQueue string `env:"CS_TEMPORAL_TASKQUEUE" envDefault:"default"`
}

func ReadConfig() (Config, error) {
	var cfg Config

	if err := env.Parse(&cfg); err != nil {
		return cfg, errors.Wrap(err, "failed to parse config")
	}

	return cfg, nil
}
