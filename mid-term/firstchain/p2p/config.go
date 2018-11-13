package p2p

import (
	"firstchain/common"

	"github.com/libp2p/go-libp2p-crypto"
	ma "github.com/multiformats/go-multiaddr"

	"encoding/base64"
)

type Config struct {
	seeds         []ma.Multiaddr // Seed nodes for initialization
	routeFilePath string         // Store route table
	privKey       crypto.PrivKey // Private key of peer
	port          int            // Listener port
	maxPeers      int            // max peers count
}

func newConfig(conf *common.Config) *Config {
	config := &Config{
		routeFilePath: conf.GetString(common.RouteFilePath),
		port:          conf.GetInt(common.Port),
		maxPeers:      conf.GetInt(common.MaxPeers),
	}

	privKeyStr := conf.GetString(common.NetPrivKey)

	pkb, err := base64.StdEncoding.DecodeString(privKeyStr)
	if err != nil {
		panic("failed to decode private key")
	}

	//privKey, err := crypto.UnmarshalPrivateKey([]byte(privKeyStr))
	privKey, err := crypto.UnmarshalPrivateKey(pkb)
	if err != nil {
		panic("failed to decode private key")
	}

	config.privKey = privKey

	//config.privKey = nil

	seeds := conf.GetSlice(common.Seeds)
	for _, seed := range seeds {
		ipfsAddr, err := ma.NewMultiaddr(seed)
		if err != nil {
			log.Errorf("invalid seed address %s", seed)
			continue
		}
		config.seeds = append(config.seeds, ipfsAddr)
	}
	return config
}
