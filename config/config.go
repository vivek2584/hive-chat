package config

import (
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

const (
	UserAgent           = "hive-chat/0.0.1"
	IdentityKeyFilePath = "identity.key"
	RendezvousString    = "hive-chat"
	ListenAddrTCP       = "/ip4/0.0.0.0/tcp/4001"
	ListenAddrQUIC      = "/ip4/0.0.0.0/udp/4001/quic-v1"
)

var RelayAddrs = []string{"/ip4/18.60.44.230/tcp/4001/p2p/12D3KooWGvs7TpG37ZFmoMiydNypHeiKCwwyjx4oBLZs3GE8vDSw"} // temporary EC2 relay address

func GetStaticRelays() []peer.AddrInfo {
	var relays []peer.AddrInfo
	for _, addr := range RelayAddrs {
		ma, err := multiaddr.NewMultiaddr(addr)
		if err != nil {
			continue
		}
		pi, err := peer.AddrInfoFromP2pAddr(ma)
		if err != nil {
			continue

		}
		relays = append(relays, *pi)
	}

	return relays
}
