package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vivek2584/hive-chat/identity"
	"github.com/vivek2584/hive-chat/node"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	i := identity.New(logger)

	n, err := node.New(ctx, logger, i)

	if err != nil {
		logger.Fatal("failed to create node", zap.Error(err))
	}

	if err := n.Bootstrap(ctx); err != nil {
		logger.Fatal("failed to bootstrap node", zap.Error(err))
	}

	err = n.StartLocalDiscovery(ctx)

	if err != nil {
		logger.Fatal("failed to start local discovery", zap.Error(err))
	}

	err = n.StartGlobalDiscovery(ctx)
	if err != nil {
		logger.Fatal("failed to start global discovery", zap.Error(err))
	}

	logger.Info("Node is running. Press CTRL-C to exit.")
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	logger.Info("Shutting down...")
	cancel()

	closeCtx, closeCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer closeCancel()

	if err := n.Close(closeCtx); err != nil {
		logger.Error("failed to close node gracefully", zap.Error(err))
	}
}
