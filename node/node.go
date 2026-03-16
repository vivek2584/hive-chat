package node

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/routing"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	rcmgr "github.com/libp2p/go-libp2p/p2p/host/resource-manager"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
	"github.com/libp2p/go-libp2p/p2p/security/noise"
	"github.com/multiformats/go-multiaddr"
	"github.com/vivek2584/hive-chat/config"
	"github.com/vivek2584/hive-chat/identity"
	"go.uber.org/zap"
)

type Node struct {
	log  *zap.Logger
	id   *identity.Identity
	host host.Host
	idht *dht.IpfsDHT
	mdns mdns.Service
}

func New(ctx context.Context, log *zap.Logger, id *identity.Identity) (*Node, error) {

	publicIP := getPublicIP()
	if publicIP != "" {
		log.Info("detected public IP", zap.String("ip", publicIP))
	}

	var idht *dht.IpfsDHT
	connManager, err := connmgr.NewConnManager(50, 100, connmgr.WithGracePeriod(time.Minute))

	if err != nil {
		log.Error("failed to create connection manager", zap.Error(err))
		return nil, err
	}

	resourceManager, err := rcmgr.NewResourceManager(rcmgr.NewFixedLimiter(rcmgr.DefaultLimits.AutoScale()))
	if err != nil {
		log.Error("failed to create resource manager", zap.Error(err))
		return nil, err
	}

	privKey, err := id.LoadIdentity(config.IdentityKeyFilePath)

	if err != nil {
		log.Error("failed to load identity key", zap.Error(err))
		return nil, err
	}

	h, err := libp2p.New(
		libp2p.Identity(privKey),
		libp2p.UserAgent(config.UserAgent),
		libp2p.ListenAddrStrings(config.ListenAddrTCP, config.ListenAddrQUIC),
		libp2p.ConnectionManager(connManager),
		libp2p.ResourceManager(resourceManager),
		libp2p.Security(noise.ID, noise.New),
		libp2p.AddrsFactory(func(addrs []multiaddr.Multiaddr) []multiaddr.Multiaddr {
			if publicIP == "" {
				return addrs
			}
			pubTCP, err1 := multiaddr.NewMultiaddr(
				fmt.Sprintf("/ip4/%s/tcp/4001", publicIP),
			)
			pubQUIC, err2 := multiaddr.NewMultiaddr(
				fmt.Sprintf("/ip4/%s/udp/4001/quic-v1", publicIP),
			)

			if err1 != nil || err2 != nil {
				return addrs
			}
			return append(addrs, pubTCP, pubQUIC)
		}),
		libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
			mode := dht.Mode(dht.ModeAuto)  // change to dht.ModeServer to configure DHT in server mode
			idht, err = dht.New(ctx, h, mode)
			return idht, err
		}),
		libp2p.NATPortMap(),
		libp2p.EnableNATService(),
		libp2p.EnableRelay(),
		libp2p.EnableAutoRelayWithStaticRelays(config.GetStaticRelays()),
		libp2p.EnableHolePunching(),
		// libp2p.ForceReachabilityPublic(),      // relay node config to force the node as publicly reachable
	)

	if err != nil {
		log.Error("failed to create libp2p host", zap.Error(err))
		return nil, err
	}

	log.Info("host created", zap.String("peerID", h.ID().String()), zap.Any("addresses", h.Addrs()))

	return &Node{
		log:  log.Named("node"),
		id:   id,
		host: h,
		idht: idht,
	}, nil
}

func (n *Node) Close(ctx context.Context) error {
	if n.mdns != nil {
		if err := n.mdns.Close(); err != nil {
			n.log.Error("failed to close mdns", zap.Error(err))
		}
	}
	if n.idht != nil {
		if err := n.idht.Close(); err != nil {
			n.log.Error("failed to close idht", zap.Error(err))
		}
	}
	if n.host != nil {
		if err := n.host.Close(); err != nil {
			n.log.Error("failed to close host", zap.Error(err))
			return err
		}
	}
	return nil
}

func getPublicIP() string {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := client.Get("https://api.ipify.org")
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	ip, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}

	publicIP := strings.TrimSpace(string(ip))
	if net.ParseIP(publicIP) == nil {
		return ""
	}

	return publicIP
}
