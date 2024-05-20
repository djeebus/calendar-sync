package clients

import (
	"calendar-sync/pkg"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
	"os"
)

func ReadConfig(filename string) (*oauth2.Config, error) {
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

	return &oauth2.Config{
		ClientID:     downloadedConfig.Web.ClientID,
		ClientSecret: downloadedConfig.Web.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:       downloadedConfig.Web.AuthURI,
			DeviceAuthURL: "",
			TokenURL:      downloadedConfig.Web.TokenURI,
			AuthStyle:     0,
		},
		RedirectURL: fmt.Sprintf("http://localhost:%d/auth/end", pkg.ListenPort),
		Scopes: []string{
			"https://www.googleapis.com/auth/calendar",
		},
	}, nil
}

func GetClient(ctx context.Context, config *oauth2.Config, tokens *oauth2.Token) (*calendar.Service, error) {
	client := config.Client(ctx, tokens)
	return calendar.NewService(ctx, option.WithHTTPClient(client))
}
