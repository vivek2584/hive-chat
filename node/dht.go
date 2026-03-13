package node

import (
	"context"
	"go.uber.org/zap"
)

func (n *Node) Bootstrap(ctx context.Context) {
	go func() {
		err := n.idht.Bootstrap(ctx)

		if err != nil {
			n.log.Error("failed to bootstrap DHT", zap.Error(err))
			return
		}

		n.log.Info("Bootstrap started")
	}()
}
