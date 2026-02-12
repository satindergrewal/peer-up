package auth

import (
	"fmt"
	"log"
	"sync"

	"github.com/libp2p/go-libp2p/core/control"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

// AuthorizedPeerGater implements the ConnectionGater interface
// It blocks connections from peers that are not in the authorized list
type AuthorizedPeerGater struct {
	authorizedPeers map[peer.ID]bool
	mu              sync.RWMutex
	logger          *log.Logger
}

// NewAuthorizedPeerGater creates a new connection gater with the given authorized peers
func NewAuthorizedPeerGater(authorizedPeers map[peer.ID]bool, logger *log.Logger) *AuthorizedPeerGater {
	if logger == nil {
		logger = log.New(log.Writer(), "[AUTH] ", log.LstdFlags)
	}
	return &AuthorizedPeerGater{
		authorizedPeers: authorizedPeers,
		logger:          logger,
	}
}

// InterceptPeerDial is called when dialing a peer
func (g *AuthorizedPeerGater) InterceptPeerDial(p peer.ID) bool {
	// Allow outbound connections to anyone
	// This is important for DHT, relay connections, etc.
	return true
}

// InterceptAddrDial is called when dialing an address
func (g *AuthorizedPeerGater) InterceptAddrDial(id peer.ID, ma multiaddr.Multiaddr) bool {
	// Allow outbound connections
	return true
}

// InterceptAccept is called when accepting a connection (before crypto handshake)
func (g *AuthorizedPeerGater) InterceptAccept(cm network.ConnMultiaddrs) bool {
	// Allow all at this stage - we'll check after crypto handshake in InterceptSecured
	return true
}

// InterceptSecured is called after the crypto handshake (peer ID is verified)
// This is the PRIMARY authorization check point
func (g *AuthorizedPeerGater) InterceptSecured(dir network.Direction, p peer.ID, addr network.ConnMultiaddrs) bool {
	g.mu.RLock()
	defer g.mu.RUnlock()

	// Only check authorization for inbound connections
	if dir == network.DirInbound {
		authorized := g.authorizedPeers[p]
		if !authorized {
			g.logger.Printf("DENIED inbound connection from unauthorized peer: %s (from %s)",
				p.String()[:16]+"...", addr.RemoteMultiaddr())
			return false
		}
		g.logger.Printf("ALLOWED inbound connection from authorized peer: %s", p.String()[:16]+"...")
	}

	// Always allow outbound connections
	return true
}

// InterceptUpgraded is called after connection upgrade (after muxer negotiation)
func (g *AuthorizedPeerGater) InterceptUpgraded(conn network.Conn) (bool, control.DisconnectReason) {
	// No additional checks needed at this stage
	return true, 0
}

// UpdateAuthorizedPeers updates the authorized peers list (for hot-reload support)
func (g *AuthorizedPeerGater) UpdateAuthorizedPeers(authorizedPeers map[peer.ID]bool) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.authorizedPeers = authorizedPeers
	g.logger.Printf("Updated authorized peers list (%d peers)", len(authorizedPeers))
}

// GetAuthorizedPeersCount returns the number of authorized peers
func (g *AuthorizedPeerGater) GetAuthorizedPeersCount() int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return len(g.authorizedPeers)
}

// IsAuthorized checks if a peer is authorized
func (g *AuthorizedPeerGater) IsAuthorized(p peer.ID) bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.authorizedPeers[p]
}

// PrintAuthorizedPeers prints the list of authorized peers (for debugging)
func (g *AuthorizedPeerGater) PrintAuthorizedPeers() {
	g.mu.RLock()
	defer g.mu.RUnlock()
	fmt.Println("Authorized peers:")
	for p := range g.authorizedPeers {
		fmt.Printf("  - %s\n", p.String())
	}
}
