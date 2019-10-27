package util

import (
	"crypto/rand"
	"io/ioutil"
	"os"

	ic "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/textileio/go-textile-core/crypto/symmetric"
	"github.com/textileio/go-textile-core/thread"
	tserv "github.com/textileio/go-textile-core/threadservice"
)

// CreateLog creates a new log with the given peer as host.
func CreateLog(host peer.ID) (info thread.LogInfo, err error) {
	sk, pk, err := ic.GenerateEd25519Key(rand.Reader)
	if err != nil {
		return
	}
	id, err := peer.IDFromPublicKey(pk)
	if err != nil {
		return
	}
	rk, err := symmetric.CreateKey()
	if err != nil {
		return
	}
	fk, err := symmetric.CreateKey()
	if err != nil {
		return
	}
	pro := ma.ProtocolWithCode(ma.P_P2P).Name
	addr, err := ma.NewMultiaddr("/" + pro + "/" + host.String())
	if err != nil {
		return
	}
	return thread.LogInfo{
		ID:        id,
		PubKey:    pk,
		PrivKey:   sk,
		ReadKey:   rk,
		FollowKey: fk,
		Addrs:     []ma.Multiaddr{addr},
	}, nil
}

// GetOrCreateLog returns the log with the given thread and log id.
// If no log exists, a new one is created under the given thread.
func GetOrCreateLog(t tserv.Threadservice, id thread.ID, lid peer.ID) (info thread.LogInfo, err error) {
	info, err = t.Store().LogInfo(id, lid)
	if err != nil {
		return
	}
	if info.PubKey != nil {
		return
	}
	info, err = CreateLog(t.Host().ID())
	if err != nil {
		return
	}
	err = t.Store().AddLog(id, info)
	return
}

// GetOwnLoad returns the log owned by the host under the given thread.
func GetOwnLog(t tserv.Threadservice, id thread.ID) (info thread.LogInfo, err error) {
	logs, err := t.Store().LogsWithKeys(id)
	if err != nil {
		return info, err
	}
	for _, lid := range logs {
		sk, err := t.Store().PrivKey(id, lid)
		if err != nil {
			return info, err
		}
		if sk != nil {
			li, err := t.Store().LogInfo(id, lid)
			if err != nil {
				return info, err
			}
			return li, nil
		}
	}
	return info, nil
}

// GetOrCreateOwnLoad returns the log owned by the host under the given thread.
// If no log exists, a new one is created under the given thread.
func GetOrCreateOwnLog(t tserv.Threadservice, id thread.ID) (info thread.LogInfo, err error) {
	info, err = GetOwnLog(t, id)
	if err != nil {
		return info, err
	}
	if info.PubKey != nil {
		return
	}
	info, err = CreateLog(t.Host().ID())
	if err != nil {
		return
	}
	err = t.Store().AddLog(id, info)
	return
}

// LoadKey at path.
func LoadKey(pth string) ic.PrivKey {
	var priv ic.PrivKey
	_, err := os.Stat(pth)
	if os.IsNotExist(err) {
		priv, _, err = ic.GenerateKeyPair(ic.Ed25519, 0)
		if err != nil {
			panic(err)
		}
		key, err := ic.MarshalPrivateKey(priv)
		if err != nil {
			panic(err)
		}
		if err = ioutil.WriteFile(pth, key, 0400); err != nil {
			panic(err)
		}
	} else if err != nil {
		panic(err)
	} else {
		key, err := ioutil.ReadFile(pth)
		if err != nil {
			panic(err)
		}
		priv, err = ic.UnmarshalPrivateKey(key)
		if err != nil {
			panic(err)
		}
	}
	return priv
}

// ParseBootstrapPeers returns address info from a list of string addresses.
func ParseBootstrapPeers(addrs []string) ([]peer.AddrInfo, error) {
	maddrs := make([]ma.Multiaddr, len(addrs))
	for i, addr := range addrs {
		var err error
		maddrs[i], err = ma.NewMultiaddr(addr)
		if err != nil {
			return nil, err
		}
	}
	return peer.AddrInfosFromP2pAddrs(maddrs...)
}
