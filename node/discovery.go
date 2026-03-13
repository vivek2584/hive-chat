package node

import (
	"context"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	"github.com/vivek2584/hive-chat/config"
	"go.uber.org/zap"
)

type discoveryNotifee struct {
	PeerChan chan peer.AddrInfo
}

func (d *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	d.PeerChan <- pi
}

func (n *Node) initMDNS() (chan peer.AddrInfo, error) {
	notifee := &discoveryNotifee{}
	notifee.PeerChan = make(chan peer.AddrInfo, 10)

	ser := mdns.NewMdnsService(n.host, config.RendezvousString, notifee)

	if err := ser.Start(); err != nil {
		n.log.Error("failed to start mdns service", zap.Error(err))
		return nil, err
	}
	return notifee.PeerChan, nil
}

func (n *Node) StartMDNSDiscovery(ctx context.Context) error {
	peerChan, err := n.initMDNS()

	if err != nil {
		n.log.Error("failed to initialize peer discovery channel", zap.Error(err))
		return err
	}

	go func() {
		for {
			select {
			case p := <-peerChan:
				if p.ID == n.host.ID() {
					continue
				}
				n.log.Info("found peer", zap.String("id", p.ID.String()), zap.Any("addresses", p.Addrs))

				if err := n.host.Connect(ctx, p); err != nil {
					n.log.Error("failed to connect to peer", zap.String("id", p.ID.String()), zap.Any("addresses", p.Addrs), zap.Error(err))
					continue
				}

			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

func (n* Node) StartDHTDiscovery(ctx context.Context) error{
	return nil
}
