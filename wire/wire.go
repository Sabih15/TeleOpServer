//go:build wireinject

// wire.go is the injection point. Run 'go run github.com/google/wire/cmd/wire ./wire/...'
// to regenerate wire_gen.go. Do not call InitializeApp from this file directly.
package wire

import (
	"github.com/google/wire"
	"github.com/sabih15/TeleOpServer/internal/config"
	"github.com/sabih15/TeleOpServer/internal/database"
	"github.com/sabih15/TeleOpServer/internal/domain/user"
	"github.com/sabih15/TeleOpServer/internal/server"
)

func InitializeApp() (*server.Server, error) {
	wire.Build(
		config.Load,
		database.NewPostgres,
		user.NewRepository,
		user.NewService,
		user.NewHandler,
		server.NewServer,
	)
	return nil, nil
}
