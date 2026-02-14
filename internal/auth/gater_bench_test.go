package auth

import (
	"testing"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
)

func BenchmarkInterceptSecuredAllowed(b *testing.B) {
	allowed := genPeerID(b)
	g := NewAuthorizedPeerGater(map[peer.ID]bool{allowed: true})
	cm := testConnMultiaddrs()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.InterceptSecured(network.DirInbound, allowed, cm)
	}
}

func BenchmarkInterceptSecuredDenied(b *testing.B) {
	denied := genPeerID(b)
	g := NewAuthorizedPeerGater(map[peer.ID]bool{})
	cm := testConnMultiaddrs()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.InterceptSecured(network.DirInbound, denied, cm)
	}
}

func BenchmarkIsAuthorized(b *testing.B) {
	p := genPeerID(b)
	g := NewAuthorizedPeerGater(map[peer.ID]bool{p: true})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.IsAuthorized(p)
	}
}
