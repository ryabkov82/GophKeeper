package main

import (
	"context"
	"log"

	"github.com/ryabkov82/gophkeeper/internal/client/app"
	"github.com/ryabkov82/gophkeeper/internal/client/config"
	"github.com/ryabkov82/gophkeeper/internal/client/tui"
	"github.com/ryabkov82/gophkeeper/internal/client/tuiiface"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	progFactory := tui.DefaultProgramFactory

	err = app.RunWithServices(cfg,
		func(ctx context.Context, services *app.AppServices, _ tuiiface.ProgramFactory) error {
			// Передаём progFactory и сигналы в tui.Run
			return tui.Run(ctx, services, progFactory)
		},
	)

	if err != nil {
		log.Fatal(err)
	}
}
