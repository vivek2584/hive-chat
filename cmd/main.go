package main

import (
	"context"
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

}
