package container

import (
	"calendar-sync/pkg"
	"calendar-sync/pkg/logs"
	"calendar-sync/pkg/persistence"
	"calendar-sync/pkg/tracing"
	"context"
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/workflow"
	"golang.org/x/oauth2"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
	"os"
)

type Container struct {
	Config         pkg.Config
	Database       *persistence.Database
	OAuth2Config   *oauth2.Config
	Logger         zerolog.Logger
	TemporalClient client.Client
}

func (c Container) Close() {
	c.Database.Close()
}

func New(ctx context.Context, cfg pkg.Config) (Container, error) {
	var err error

	ctr := Container{
		Config: cfg,
		Logger: logs.New(cfg),
	}

	ctr.Database, err = persistence.NewDatabase(ctx, cfg)
	if err != nil {
		return ctr, err
	}

	ctr.TemporalClient, err = client.DialContext(ctx, client.Options{
		HostPort:  cfg.TemporalHostPort,
		Namespace: cfg.TemporalNamespace,
		Identity:  cfg.TemporalIdentity,
		Logger:    logs.NewTemporalLogger(ctr.Logger),
		ContextPropagators: []workflow.ContextPropagator{
			tracing.NewCorrelationIDPropagator(),
		},
	})
	if err != nil {
		return ctr, errors.Wrap(err, "failed to dial the temporal server")
	}

	ctr.OAuth2Config, err = readConfig(cfg.ClientSecretsPath)
	if err != nil {
		return ctr, errors.Wrap(err, "failed to read client secrets")
	}

	return ctr, nil
}

func (c Container) GetCalendarClient(ctx context.Context) (*calendar.Service, error) {
	tokens, err := c.Database.GetTokens(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get tokens")
	}

	if tokens.RefreshToken != "" {
		if !tokens.Valid() {
			log.Info().Msg("token is invalid, refreshing")
			ts := c.OAuth2Config.TokenSource(ctx, tokens)
			tokens, err = ts.Token()
			if err != nil {
				return nil, errors.Wrap(err, "failed to refresh token")
			}

			if err = c.Database.UpdateTokens(ctx, tokens); err != nil {
				return nil, errors.Wrap(err, "failed to store updated tokens")
			}

		}
	}

	return c.GetCalendarClientWithToken(ctx, tokens)
}

func (c Container) GetCalendarClientWithToken(ctx context.Context, tokens *oauth2.Token) (*calendar.Service, error) {
	oauth2client := c.OAuth2Config.Client(ctx, tokens)
	cal, err := calendar.NewService(ctx, option.WithHTTPClient(oauth2client))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create calendar")
	}

	return cal, nil
}

func readConfig(filename string) (*oauth2.Config, error) {
	var downloadedConfig struct {
		Web struct {
			ClientID                string   `json:"client_id"`
			ProjectID               string   `json:"project_id"`
			AuthURI                 string   `json:"auth_uri"`
			TokenURI                string   `json:"token_uri"`
			AuthProviderX509CertUrl string   `json:"auth_provider_x509_cert_url"`
			ClientSecret            string   `json:"client_secret"`
			RedirectUris            []string `json:"redirect_uris"`
		} `json:"web"`
	}

	clientSecret, err := os.ReadFile(filename)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read secrets")
	}

	if err := json.Unmarshal(clientSecret, &downloadedConfig); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal secrets")
	}

	model := oauth2.Config{
		ClientID:     downloadedConfig.Web.ClientID,
		ClientSecret: downloadedConfig.Web.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:       downloadedConfig.Web.AuthURI,
			DeviceAuthURL: "",
			TokenURL:      downloadedConfig.Web.TokenURI,
			AuthStyle:     0,
		},
		Scopes: []string{
			"https://www.googleapis.com/auth/calendar",
		},
	}

	for _, redirectUri := range downloadedConfig.Web.RedirectUris {
		model.RedirectURL = redirectUri
	}

	return &model, nil
}
