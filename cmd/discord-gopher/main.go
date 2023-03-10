package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/bsponge/discordGopher/pkg/client"
	"github.com/bsponge/discordGopher/pkg/log"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	log.Logger().Info("Hello")

	c, err := client.NewClient()
	if err != nil {
		log.Logger().WithError(err).Error("Failed to create the client")
		return
	}

	err = c.Start(ctx)
	if err != nil {
		log.Logger().WithError(err).Error("Failed to start the client")
		return
	}
	defer c.Stop()

	<-ctx.Done()
	log.Logger().Info("Main context canceled. Exiting...")
}
