package node

import (
	"context"
	"sync"
	"time"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/peer"
	"go.uber.org/zap"
)

func (n *Node) Bootstrap(ctx context.Context) {
	n.log.Info("connecting to default bootstrap peers...")
	var wg sync.WaitGroup
	for _, peerAddr := range dht.DefaultBootstrapPeers {
		peerinfo, err := peer.AddrInfoFromP2pAddr(peerAddr)
		if err != nil {
			continue
		}
		wg.Add(1)
		go func(pi *peer.AddrInfo) {
			defer wg.Done()
			if err := n.host.Connect(ctx, *pi); err != nil {
				n.log.Warn("failed to connect to bootstrap peer", zap.String("peer", pi.ID.String()), zap.Error(err))
			} else {
				n.log.Info("connected to bootstrap peer", zap.String("peer", pi.ID.String()))
			}
		}(peerinfo)
	}
	wg.Wait()

	go func() {
		err := n.idht.Bootstrap(ctx)

		if err != nil {
			n.log.Error("failed to bootstrap DHT", zap.Error(err))
			return
		}
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
