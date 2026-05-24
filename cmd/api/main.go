// @title           TeleOpServer API
// @version         1.0
// @description     TeleOpServer backend API
// @host            localhost:8080
// @BasePath        /
// @securityDefinitions.apikey BearerAuth
// @in              header
// @name            Authorization
package main

import (
	"os"
	"os/exec"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/sabih15/TeleOpServer/wire"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	app, err := wire.InitializeApp()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize application")
	}

	log.Info().Msg("TeleOpServer starting")
	go openBrowser("http://localhost:8080")
	if err := app.Run(); err != nil {
		log.Fatal().Err(err).Msg("server stopped")
	}
}

func openBrowser(url string) {
	exec.Command("cmd", "/c", "start", url).Start()
}
