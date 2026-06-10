//go:build wireinject

// wire.go is the injection point. Run 'go run github.com/google/wire/cmd/wire ./wire/...'
// to regenerate wire_gen.go. Do not call InitializeApp from this file directly.
package wire

import (
	"github.com/google/wire"
	tocommands "github.com/sabih15/TeleOpServer/internal/modules/TOCommands"
	"github.com/sabih15/TeleOpServer/internal/modules/user"
	"github.com/sabih15/TeleOpServer/internal/platform/config"
	"github.com/sabih15/TeleOpServer/internal/platform/database"
	"github.com/sabih15/TeleOpServer/internal/platform/server"
)

func InitializeApp() (*server.Server, error) {
	wire.Build(
		config.Load,
		database.NewPostgres,
		user.ProviderSet,
		tocommands.ProviderSet,
		provideRouter,
		server.NewServer,
	)
	return nil, nil
}
