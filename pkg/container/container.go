package container

import (
	"calendar-sync/pkg/clients"
	"calendar-sync/pkg/persistence"
	"context"
	"github.com/pkg/errors"
	"go.temporal.io/sdk/client"
	"golang.org/x/oauth2"
)

type Container struct {
	Database       *persistence.Database
	OAuth2Config   *oauth2.Config
	TemporalClient client.Client
	TaskQueue      string
}

func New(ctx context.Context) (Container, func(), error) {
	var (
		ctr Container
		err error
	)

	var cleaners []func()

	ctr.TaskQueue = "default"

	ctr.Database, err = persistence.NewDatabase()
	if err != nil {
		return ctr, nil, err
	}
	cleaners = append(cleaners, ctr.Database.Close)

	ctr.TemporalClient, err = client.DialContext(ctx, client.Options{})
	if err != nil {
		return ctr, nil, errors.Wrap(err, "failed to dial the temporal server")
	}

	ctr.OAuth2Config = &clients.Config

	return ctr, combineCleaners(cleaners), nil
}

func combineCleaners(cleaners []func()) func() {
	return func() {
		for _, cleaner := range cleaners {
			cleaner()
		}
	}
}
