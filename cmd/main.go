package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/vivek2584/hive-chat/identity"
	"github.com/vivek2584/hive-chat/node"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	ctx := context.Background()

	i := identity.New(logger)

	n, err := node.New(ctx, logger, i)

	if err != nil{
		return
	}

	n.Bootstrap(ctx)
	n.WaitForBootstrap(ctx)

	err = n.StartLocalDiscovery(ctx)

	if err != nil{
		return
	}

	err = n.StartGlobalDiscovery(ctx)
	if err != nil{
		return
	}

	logger.Info("Node is running. Press CTRL-C to exit.")
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	logger.Info("Shutting down...")
}
