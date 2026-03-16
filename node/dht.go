package node

import (
	"context"
	"sync"

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
