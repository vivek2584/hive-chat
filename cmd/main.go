package main

import (
	"github.com/vivek2584/hive-chat/identity"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()
	
	i := identity.New(logger)

}
