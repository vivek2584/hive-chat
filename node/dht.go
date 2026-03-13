package node

import (
	"context"
	"time"

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

func (n *Node) WaitForBootstrap(ctx context.Context) {
	for {
		select {

		case <-ctx.Done():
			return

		default:
			if n.idht.RoutingTable().Size() > 0 {
				n.log.Info("DHT bootstrap complete",
					zap.Int("peers", n.idht.RoutingTable().Size()),
				)
				return
			}
			n.log.Info("waiting for DHT bootstrap...")
			time.Sleep(time.Second)
		}
	}
}
