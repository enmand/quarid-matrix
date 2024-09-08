package matrix

import (
	"context"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog"
	"github.com/samber/mo"
	"go.mau.fi/util/dbutil"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/crypto/cryptohelper"
)

type Option[S Store] func(*options[S])

type e2eeOpts[S Store] struct {
	Key      string
	Password string
	Store    S
}

type options[S Store] struct {
	homeserver string

	e2ee mo.Option[e2eeOpts[S]]

	logger mo.Option[zerolog.Logger]
}

// Client is a Matrix client based on mautrix/go.
type Client[S Store] struct {
	c *mautrix.Client

	options *options[S]
}

type Store interface {
	~string | *dbutil.Database
}

// WithE2EE enables end-to-end encryption, and requires the encryption key.
func WithE2EE[S Store](key, password string, store S) Option[S] {
	return func(o *options[S]) {
		o.e2ee = mo.Some(e2eeOpts[S]{
			Key:      key,
			Password: password,
			Store:    store,
		})
	}
}

func WithLogger[S Store](logger zerolog.Logger) Option[S] {
	return func(o *options[S]) {
		o.logger = mo.Some(logger)
	}
}

// NewClient creates a new Matrix client.
func NewClient[S Store](homeserver, userID string, opts ...Option[S]) (*Client[S], error) {
	o := &options[S]{
		homeserver: homeserver,
	}

	for _, opt := range opts {
		opt(o)
	}

	client, err := mautrix.NewClient(o.homeserver, "", "")
	if err != nil {
		return nil, fmt.Errorf("unable to create to Matrix client: %w", err)
	}

	fmt.Printf("%# v", o)

	client.Log = o.logger.OrElse(zerolog.Nop())

	if e2ee, ok := o.e2ee.Get(); ok {
		ch, err := cryptohelper.NewCryptoHelper(client, []byte(e2ee.Key), e2ee.Store)
		if err != nil {
			return nil, fmt.Errorf("unable to create crypto helper: %w", err)
		}

		ch.LoginAs = &mautrix.ReqLogin{
			Type: mautrix.AuthTypePassword,
			Identifier: mautrix.UserIdentifier{
				Type: mautrix.IdentifierTypeUser,
				User: userID,
			},
			Password:         e2ee.Password,
			StoreCredentials: true,
		}

		client.Crypto = ch
	}

	return &Client[S]{
		client,

		o,
	}, nil
}

// Sync starts the sync loop for the client.
func (c *Client[S]) Sync(ctx context.Context) error {
	if err := c.c.Crypto.Init(ctx); err != nil {
		return fmt.Errorf("unable to initialize crypto: %w", err)
	}

	return c.c.SyncWithContext(ctx)
}
