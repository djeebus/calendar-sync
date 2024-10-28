package container

import (
	"calendar-sync/pkg/persistence"
	"context"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
)

type tokenPersistor struct {
	ctx      context.Context
	db       *persistence.Database
	original *oauth2.Token
	next     oauth2.TokenSource
}

var _ oauth2.TokenSource = new(tokenPersistor)

func (t *tokenPersistor) Token() (*oauth2.Token, error) {
	if t.original != nil && t.original.Valid() {
		log.Debug().Msg("returning original tokens")
		return t.original, nil
	}

	log.Debug().Msg("creating new tokens")
	tok, err := t.next.Token()
	if err != nil {
		return nil, errors.Wrap(err, "t.next.Token() failed")
	}

	log.Debug().Msg("storing new tokens")
	if err := t.db.SetTokens(t.ctx, tok); err != nil {
		return nil, errors.Wrap(err, "t.db.SetTokens() failed")
	}

	log.Debug().Msg("returning new tokens")
	return tok, nil
}

func newTokenPersistor(db *persistence.Database, tokens oauth2.TokenSource) *tokenPersistor {
	return &tokenPersistor{
		db:   db,
		next: tokens,
	}
}
