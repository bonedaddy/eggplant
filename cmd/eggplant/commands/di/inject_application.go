package di

import (
	"path/filepath"

	"github.com/boreq/eggplant/config"
	"github.com/boreq/eggplant/pkg/service/adapters"
	authAdapters "github.com/boreq/eggplant/pkg/service/adapters/auth"
	"github.com/boreq/eggplant/pkg/service/application"
	"github.com/boreq/eggplant/pkg/service/application/auth"
	"github.com/boreq/eggplant/pkg/service/application/queries"
	"github.com/google/wire"
	bolt "go.etcd.io/bbolt"
)

//lint:ignore U1000 because
var appSet = wire.NewSet(
	wire.Struct(new(application.Application), "*"),

	wire.Struct(new(application.Auth), "*"),
	auth.NewRegisterInitialHandler,
	auth.NewLoginHandler,
	auth.NewLogoutHandler,
	auth.NewCheckAccessTokenHandler,

	wire.Struct(new(application.Commands), "*"),

	wire.Struct(new(application.Queries), "*"),
	queries.NewStatsHandler,

	wire.Bind(new(queries.UserRepository), new(*authAdapters.UserRepository)),
	wire.Bind(new(auth.UserRepository), new(*authAdapters.UserRepository)),
	authAdapters.NewUserRepository,

	wire.Bind(new(authAdapters.PasswordHasher), new(*authAdapters.BcryptPasswordHasher)),
	authAdapters.NewBcryptPasswordHasher,

	wire.Bind(new(authAdapters.AccessTokenGenerator), new(*authAdapters.CryptoAccessTokenGenerator)),
	authAdapters.NewCryptoAccessTokenGenerator,

	newBolt,
)

func newBolt(conf *config.Config) (*bolt.DB, error) {
	path := filepath.Join(conf.DataDirectory, "eggplant.database")
	return adapters.NewBolt(path)
}