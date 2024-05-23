package container

import (
	"calendar-sync/pkg"
	"calendar-sync/pkg/clients"
	"calendar-sync/pkg/persistence"
	"context"
	"github.com/pkg/errors"
	"go.temporal.io/sdk/client"
	"golang.org/x/oauth2"
)

type Container struct {
	Config         pkg.Config
	Database       *persistence.Database
	OAuth2Config   *oauth2.Config
	TemporalClient client.Client
}

func (c Container) Close() {
	c.Database.Close()
}

func New(ctx context.Context, cfg pkg.Config) (Container, error) {
	var err error

	ctr := Container{
		Config: cfg,
	}

	ctr.Database, err = persistence.NewDatabase(ctx, cfg)
	if err != nil {
		return ctr, err
	}

	ctr.TemporalClient, err = client.DialContext(ctx, client.Options{
		HostPort:  cfg.TemporalHostPort,
		Namespace: cfg.TemporalNamespace,
		Identity:  cfg.TemporalIdentity,
	})
	if err != nil {
		return ctr, errors.Wrap(err, "failed to dial the temporal server")
	}

	ctr.OAuth2Config, err = clients.ReadConfig(cfg.ClientSecretsPath)
	if err != nil {
		return ctr, errors.Wrap(err, "failed to read client secrets")
	}

	return ctr, nil
}
