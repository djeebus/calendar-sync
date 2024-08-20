package pkg

import (
	"github.com/caarlos0/env/v11"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"reflect"
	"time"
)

type Config struct {
	OwnerEmailAddress string `env:"CS_OWNER_EMAIL_ADDRESS,required"`

	LogJson  bool          `env:"CS_LOG_JSON" envDefault:"true"`
	LogLevel zerolog.Level `env:"CS_LOG_LEVEL" envDefault:"INFO"`

	ClientSecretsPath string `env:"CS_CLIENT_SECRETS_PATH,required"`
	Listen            string `env:"CS_LISTEN" envDefault:":31425"`
	WebhookUrl        string `env:"CS_WEBHOOK_URL"`

	DatabaseDriver string `env:"CS_DATABASE_DRIVER" envDefault:"sqlite3"`
	DatabaseSource string `env:"CS_DATABASE_SOURCE" envDefault:"./database.db"`

	TemporalHostPort  string `env:"CS_TEMPORAL_HOSTPORT"`
	TemporalNamespace string `env:"CS_TEMPORAL_NAMESPACE"`
	TemporalIdentity  string `env:"CS_TEMPORAL_IDENTITY"`
	TemporalTaskQueue string `env:"CS_TEMPORAL_TASKQUEUE" envDefault:"default"`

	JwtAlgorithm string        `env:"JWT_ALGORITHM" envDefault:"HS256"`
	JwtDuration  time.Duration `env:"JWT_DURATION" envDefault:"24h"`
	JwtIssuer    string        `env:"JWT_ISSUER" envDefault:"calendar-sync-web"`
	JwtSecretKey string        `env:"JWT_SECRET_KEY,required"`
}

func ReadConfig() (Config, error) {
	var cfg Config

	opts := env.Options{
		FuncMap: map[reflect.Type]env.ParserFunc{
			reflect.TypeOf(zerolog.DebugLevel): func(v string) (interface{}, error) {
				return zerolog.ParseLevel(v)
			},
		},
	}

	if err := env.ParseWithOptions(&cfg, opts); err != nil {
		return cfg, errors.Wrap(err, "failed to parse config")
	}

	return cfg, nil
}
