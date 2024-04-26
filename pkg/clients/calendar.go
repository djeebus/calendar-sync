package clients

import (
	"calendar-sync/pkg"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

//go:embed client_secret.json
var clientSecret []byte

var Config oauth2.Config

func init() {
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

	if err := json.Unmarshal(clientSecret, &downloadedConfig); err != nil {
		panic("failed to deserialize")
	}

	Config = oauth2.Config{
		ClientID:     downloadedConfig.Web.ClientID,
		ClientSecret: downloadedConfig.Web.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:       downloadedConfig.Web.AuthURI,
			DeviceAuthURL: "",
			TokenURL:      downloadedConfig.Web.TokenURI,
			AuthStyle:     0,
		},
		RedirectURL: fmt.Sprintf("http://localhost:%d", pkg.ListenPort),
		Scopes: []string{
			"https://www.googleapis.com/auth/calendar",
		},
	}
}

func GetClient(ctx context.Context, tokens *oauth2.Token) (*calendar.Service, error) {
	client := Config.Client(ctx, tokens)
	return calendar.NewService(ctx, option.WithHTTPClient(client))
}
