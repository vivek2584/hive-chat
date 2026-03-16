package node

import (
	"context"
	"time"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	"github.com/vivek2584/hive-chat/config"
	"go.uber.org/zap"
)

type discoveryNotifee struct {
	PeerChan chan peer.AddrInfo
}

func (d *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	select {
	case d.PeerChan <- pi:
	default:
	}
}

func (n *Node) initMDNS() (chan peer.AddrInfo, error) {
	notifee := &discoveryNotifee{}
	notifee.PeerChan = make(chan peer.AddrInfo, 10)

	ser := mdns.NewMdnsService(n.host, config.RendezvousString, notifee)

	if err := ser.Start(); err != nil {
		n.log.Error("failed to start mdns service", zap.Error(err))
		return nil, err
	}
	n.mdns = ser
	return notifee.PeerChan, nil
}

func (n *Node) StartLocalDiscovery(ctx context.Context) error {
	peerChan, err := n.initMDNS()

	if err != nil {
		n.log.Error("failed to initialize peer discovery channel", zap.Error(err))
		return err
	}

	go func() {
		for {
			select {
			case p, ok := <-peerChan:
				if !ok {
					return
				}
				if p.ID == n.host.ID() {
					continue
				}
				if n.host.Network().Connectedness(p.ID) == network.Connected {
					continue
				}

				n.log.Info("found local peer", zap.String("id", p.ID.String()), zap.Any("addresses", p.Addrs))

				go func(p peer.AddrInfo) {
					dialCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
					defer cancel()
					if err := n.host.Connect(dialCtx, p); err != nil {
						n.log.Error("failed to connect to local peer",
							zap.String("id", p.ID.String()),
							zap.Error(err),
						)
						return
					}
					n.log.Info("connected to local peer", zap.String("id", p.ID.String()), zap.Any("addresses", p.Addrs))
					n.LogConnectedPeers()
				}(p)

			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

func (n *Node) StartGlobalDiscovery(ctx context.Context) error {
	routingDiscovery := drouting.NewRoutingDiscovery(n.idht)
	advCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	_, err := routingDiscovery.Advertise(advCtx, config.RendezvousString)

	if err != nil {
		n.log.Error("failed to advertise", zap.Error(err))
		return err
	}

	n.log.Info("successfully advertised node")

	go func() {
		for {
			select {
			case <-ctx.Done():
				return

			default:
				peerChan, err := routingDiscovery.FindPeers(ctx, config.RendezvousString)

				if err != nil {
					n.log.Error("peer discovery failed", zap.Error(err))
					select {
					case <-time.After(10 * time.Second):
					case <-ctx.Done():
						return
					}
					continue
				}

				for p := range peerChan {
					if ctx.Err() != nil {
						return
					}

					if p.ID == n.host.ID() {
						continue
					}

					if n.host.Network().Connectedness(p.ID) == network.Connected {
						continue
					}

					n.log.Info("found global peer", zap.String("id", p.ID.String()), zap.Any("addresses", p.Addrs))

					go func(p peer.AddrInfo) {
						dialCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
						defer cancel()
						if err := n.host.Connect(dialCtx, p); err != nil {
							n.log.Error("failed to connect to global peer", zap.String("id", p.ID.String()), zap.Any("addresses", p.Addrs), zap.Error(err))
							return
						}
						n.log.Info("connected to global peer", zap.String("id", p.ID.String()), zap.Any("addresses", p.Addrs))
						n.LogConnectedPeers()
					}(p)

				}

				select {
				case <-time.After(10 * time.Second):
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return nil
}

func (n *Node) LogConnectedPeers() {
	peers := n.host.Network().Peers()
	n.log.Info("connected peers",
		zap.Int("count", len(peers)),
		zap.Any("peers", peers),
	)
}
