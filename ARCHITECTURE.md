# peer-up Architecture

This document describes the technical architecture of peer-up, from current implementation to future vision.

## Table of Contents

- [Current Architecture (Phase 1-3)](#current-architecture-phase-1-3)
- [Target Architecture (Phase 4+)](#target-architecture-phase-4)
- [Core Concepts](#core-concepts)
- [Security Model](#security-model)
- [Naming System](#naming-system)
- [Federation Model](#federation-model)
- [Mobile Architecture](#mobile-architecture)

---

## Current Architecture (Phase 1-3)

### Component Overview

```
peer-up/
├── relay-server/        # Circuit relay v2 (VPS)
│   └── main.go          # Relay with optional authentication
│
├── home-node/           # Service host (behind CGNAT)
│   └── main.go          # DHT advertiser, protocol responder
│
├── client-node/         # Service consumer (mobile/laptop)
│   └── main.go          # DHT searcher, protocol initiator
│
├── internal/
│   ├── config/          # YAML configuration loading
│   │   ├── config.go
│   │   └── loader.go
│   └── auth/            # SSH-style authentication
│       ├── authorized_keys.go
│       └── gater.go     # ConnectionGater implementation
│
└── cmd/
    └── keytool/         # Key management CLI
        ├── main.go
        └── commands/
```

### Network Topology (Current)

```
┌─────────────────────────────────────────────────────────────┐
│                      Internet                                │
└─────────────────────────────────────────────────────────────┘
                            │
              ┌─────────────┼─────────────┐
              │                           │
              ▼                           ▼
    ┌──────────────────┐        ┌──────────────────┐
    │   Relay Server   │        │   Client Node    │
    │      (VPS)       │        │  (Phone/Laptop)  │
    │   Public IP      │        │   CGNAT/Mobile   │
    └────────┬─────────┘        └─────────┬────────┘
             │                            │
             │ Circuit Relay v2           │
             │ (hop protocol)             │
             │                            │
             └────────────┬───────────────┘
                          │
                          ▼
                 ┌──────────────────┐
                 │    Home Node     │
                 │ (Behind Starlink)│
                 │   CGNAT + IPv6   │
                 │    Firewall      │
                 └──────────────────┘
```

**Connection Flow**:
1. Home node connects outbound to relay → makes reservation
2. Client connects outbound to relay
3. Client dials home via `/p2p-circuit` address
4. Relay bridges connection (both sides outbound-only)
5. DCUtR attempts hole-punching for direct upgrade

### Authentication Flow

```
Client Attempts Connection to Home Node
         │
         ▼
   ┌──────────────────────────────────┐
   │  libp2p Transport Handshake      │
   │  (Noise protocol, key exchange)  │
   └──────────────────┬───────────────┘
                      │
                      ▼
        ┌─────────────────────────────┐
        │  ConnectionGater.           │
        │  InterceptSecured()         │
        │                             │
        │  Check peer ID against      │
        │  authorized_keys            │
        └──────────┬──────────────────┘
                   │
         ┌─────────┴─────────┐
         │                   │
         ▼                   ▼
    ✅ Authorized      ❌ Unauthorized
    Connection         Connection
    Allowed            DENIED
         │
         ▼
   ┌──────────────────────────────────┐
   │  Protocol Handler                │
   │  (defense-in-depth check)        │
   │                                  │
   │  if !authorizer.IsAuthorized():  │
   │    close stream                  │
   └──────────────────────────────────┘
```

---

## Target Architecture (Phase 4+)

### Library-First Structure

```
peer-up/
├── pkg/p2pnet/              # 🆕 Core library (importable)
│   ├── network.go           # P2P network setup
│   ├── service.go           # Service registry
│   ├── proxy.go             # TCP↔Stream proxy
│   ├── naming.go            # Name resolution
│   └── federation.go        # Network peering
│
├── internal/                # Internal packages
│   ├── config/              # Configuration (existing)
│   ├── auth/                # Authentication (existing)
│   └── tun/                 # 🆕 TUN/TAP interface
│
├── cmd/
│   ├── gateway/             # 🆕 Multi-mode daemon
│   ├── keytool/             # Key management (existing)
│   └── peerup/              # 🆕 CLI tool
│
├── examples/                # 🆕 Example implementations
│   ├── home-node/           # Moved from root
│   ├── client-node/         # Moved from root
│   └── custom-service/      # Example: custom protocol
│
├── relay-server/            # Relay (existing)
├── mobile/                  # 🆕 Mobile apps
│   ├── ios/
│   └── android/
│
└── docs/                    # 🆕 Extended documentation
    ├── ARCHITECTURE.md      # This file
    ├── ROADMAP.md
    └── examples/
```

### Service Exposure Architecture

```
┌──────────────────────────────────────────────────────────────┐
│  Application Layer (User's Services)                         │
│  ┌────────┐  ┌────────┐  ┌────────┐  ┌────────┐            │
│  │  SSH   │  │  HTTP  │  │  SMB   │  │ Custom │            │
│  │  :22   │  │  :80   │  │  :445  │  │ :9999  │            │
│  └───┬────┘  └───┬────┘  └───┬────┘  └───┬────┘            │
└──────┼───────────┼───────────┼───────────┼─────────────────┘
       │           │           │           │
       └───────────┴───────────┴───────────┘
                   │
                   ▼
       ┌────────────────────────────┐
       │   Service Registry         │
       │   (pkg/p2pnet/service.go)  │
       │                            │
       │   "ssh"  → localhost:22    │
       │   "web"  → localhost:80    │
       │   "smb"  → localhost:445   │
       │   "custom" → localhost:9999│
       └────────────┬───────────────┘
                    │
                    ▼
       ┌────────────────────────────┐
       │   TCP ↔ Stream Proxy       │
       │   (pkg/p2pnet/proxy.go)    │
       │                            │
       │   Bidirectional relay:     │
       │   TCP socket ↔ libp2p      │
       │   stream                   │
       └────────────┬───────────────┘
                    │
                    ▼
       ┌────────────────────────────┐
       │   libp2p Network           │
       │   (with authentication)    │
       │                            │
       │   Protocol:                │
       │   /peerup/ssh/1.0.0        │
       │   /peerup/http/1.0.0       │
       │   /peerup/smb/1.0.0        │
       └────────────────────────────┘
```

### Gateway Daemon Modes

#### Mode 1: SOCKS Proxy (No Root Required)

```
┌─────────────────────────────────────────────────────────┐
│  Applications (configured to use SOCKS)                 │
│  ┌────────┐  ┌──────────┐  ┌──────────────┐           │
│  │  SSH   │  │  Browser │  │  Custom App  │           │
│  └───┬────┘  └────┬─────┘  └──────┬───────┘           │
└──────┼────────────┼────────────────┼─────────────────  ┘
       └────────────┴────────────────┘
                    │
                    ▼
       ┌────────────────────────────┐
       │   SOCKS5 Proxy             │
       │   localhost:1080           │
       │                            │
       │   Translates:              │
       │   "laptop.grewal:22"       │
       │   → peer ID                │
       │   → P2P connection         │
       └────────────┬───────────────┘
                    │
                    ▼
       ┌────────────────────────────┐
       │   P2P Network              │
       │   (pkg/p2pnet)             │
       └────────────────────────────┘
```

#### Mode 2: DNS Server

```
┌─────────────────────────────────────────────────────────┐
│  Applications (use system DNS)                          │
│  ┌────────┐  ┌──────────┐  ┌──────────────┐           │
│  │  SSH   │  │  Browser │  │  SMB Client  │           │
│  └───┬────┘  └────┬─────┘  └──────┬───────┘           │
└──────┼────────────┼────────────────┼─────────────────  ┘
       └────────────┴────────────────┘
                    │
              DNS Query:
              "laptop.grewal.p2p"
                    │
                    ▼
       ┌────────────────────────────┐
       │   Local DNS Server         │
       │   localhost:53             │
       │                            │
       │   Resolves:                │
       │   laptop.grewal.p2p        │
       │   → 10.64.1.5              │
       └────────────┬───────────────┘
                    │
                    ▼
       ┌────────────────────────────┐
       │   Virtual IP Router        │
       │                            │
       │   10.64.1.5 → peer ID      │
       │   → P2P connection         │
       └────────────────────────────┘
```

#### Mode 3: TUN/TAP Virtual Network (Requires Root)

```
┌─────────────────────────────────────────────────────────┐
│  Applications (completely transparent)                  │
│  ┌────────┐  ┌──────────┐  ┌──────────────┐           │
│  │  SSH   │  │  Browser │  │  ANY App     │           │
│  └───┬────┘  └────┬─────┘  └──────┬───────┘           │
└──────┼────────────┼────────────────┼─────────────────  ┘
       └────────────┴────────────────┘
                    │
              Normal TCP/UDP
              to 10.64.x.x
                    │
                    ▼
       ┌────────────────────────────┐
       │   Kernel Network Stack     │
       └────────────┬───────────────┘
                    │
                    ▼
       ┌────────────────────────────┐
       │   TUN Interface (peer0)    │
       │   10.64.0.1/16             │
       │                            │
       │   Intercepts all packets   │
       │   to 10.64.0.0/16          │
       └────────────┬───────────────┘
                    │
                    ▼
       ┌────────────────────────────┐
       │   Gateway Daemon           │
       │                            │
       │   Packet → Peer ID lookup  │
       │   → P2P stream             │
       │   → Forward data           │
       └────────────────────────────┘
```

---

## Core Concepts

### 1. Service Definition

Services are defined in configuration and registered at runtime:

```go
type Service struct {
    Name         string   // "ssh", "web", etc.
    Protocol     string   // "/peerup/ssh/1.0.0"
    LocalAddress string   // "localhost:22"
    Enabled      bool     // Enable/disable
}

type ServiceRegistry struct {
    services map[string]*Service
    host     host.Host
}

func (r *ServiceRegistry) RegisterService(svc *Service) error {
    // Set up stream handler for this service's protocol
    r.host.SetStreamHandler(svc.Protocol, func(s network.Stream) {
        // 1. Authorize peer
        if !r.isAuthorized(s.Conn().RemotePeer(), svc.Name) {
            s.Close()
            return
        }

        // 2. Dial local service
        localConn, err := net.Dial("tcp", svc.LocalAddress)
        if err != nil {
            s.Close()
            return
        }

        // 3. Bidirectional proxy
        go io.Copy(s, localConn)
        io.Copy(localConn, s)
    })
}
```

### 2. Bidirectional TCP↔Stream Proxy

```go
func ProxyStreamToTCP(stream network.Stream, tcpAddr string) error {
    // Connect to local TCP service
    tcpConn, err := net.Dial("tcp", tcpAddr)
    if err != nil {
        return err
    }
    defer tcpConn.Close()

    // Bidirectional copy
    errCh := make(chan error, 2)

    go func() {
        _, err := io.Copy(tcpConn, stream)
        errCh <- err
    }()

    go func() {
        _, err := io.Copy(stream, tcpConn)
        errCh <- err
    }()

    // Wait for either direction to finish
    return <-errCh
}
```

### 3. Name Resolution

```go
type NameResolver interface {
    Resolve(name string) (peer.ID, error)
}

type LocalFileResolver struct {
    names map[string]peer.ID
}

func (r *LocalFileResolver) Resolve(name string) (peer.ID, error) {
    if id, ok := r.names[name]; ok {
        return id, nil
    }
    return "", ErrNotFound
}

type DHTResolver struct {
    dht *dht.IpfsDHT
}

func (r *DHTResolver) Resolve(name string) (peer.ID, error) {
    // Query DHT for network's relay
    // Ask relay for peer name → ID mapping
    // Return peer ID
}

// Multi-tier resolution
func Resolve(name string, resolvers []NameResolver) (peer.ID, error) {
    for _, resolver := range resolvers {
        if id, err := resolver.Resolve(name); err == nil {
            return id, nil
        }
    }
    // If no resolver works, try to parse as direct peer ID
    return peer.Decode(name)
}
```

---

## Security Model

### Authentication Layers

**Layer 1: Network Level (ConnectionGater)**
- Executed during connection handshake
- Blocks unauthorized peers before any data exchange
- Fastest rejection (minimal resource usage)

**Layer 2: Protocol Level (Stream Handler)**
- Defense-in-depth validation
- Per-service authorization (optional)
- Can override global authorized_keys

### Per-Service Authorization

```yaml
# home-node.yaml
security:
  authorized_keys_file: "authorized_keys"  # Global default

services:
  ssh:
    enabled: true
    local_address: "localhost:22"
    authorized_keys: "ssh_authorized_keys"  # Override

  web:
    enabled: true
    local_address: "localhost:80"
    # Uses global authorized_keys
```

### Federation Trust Model

```yaml
# relay-server.yaml
federation:
  peers:
    - network_name: "alice"
      relay: "/ip4/.../p2p/..."
      trust_level: "full"      # Bidirectional routing

    - network_name: "bob"
      relay: "/ip4/.../p2p/..."
      trust_level: "one_way"   # Only alice → grewal, not grewal → alice
```

---

## Naming System

### Multi-Tier Resolution

```
User Request: ssh user@laptop.grewal
         │
         ▼
┌────────────────────────────────────┐
│  Tier 1: Local Override            │
│  Check: ~/.peerup/names.yaml       │
│  laptop.grewal → 12D3KooW...       │
└──────────┬─────────────────────────┘
           │ Not found
           ▼
┌────────────────────────────────────┐
│  Tier 2: Network-Scoped            │
│  Parse: laptop.grewal              │
│  Query: grewal relay for "laptop"  │
│  Response: 12D3KooW...             │
└──────────┬─────────────────────────┘
           │ Relay unreachable
           ▼
┌────────────────────────────────────┐
│  Tier 3: Blockchain (if enabled)   │
│  Query: Ethereum smart contract    │
│  grewal.register["laptop"]         │
│  Response: 12D3KooW...             │
└──────────┬─────────────────────────┘
           │ Not registered
           ▼
┌────────────────────────────────────┐
│  Tier 4: Direct Peer ID            │
│  Try: peer.Decode("laptop.grewal") │
│  Fails → Error: "Name not found"   │
└────────────────────────────────────┘
```

### Network-Scoped Name Format

```
Format: <hostname>.<network>[.<tld>]

Examples:
laptop.grewal           # Query grewal relay
desktop.alice           # Query alice relay
phone.bob.p2p           # Query bob relay (explicit .p2p TLD)
home.grewal.local       # mDNS compatible
```

---

## Federation Model

### Relay Peering

```
┌──────────────────────────────────────────────────────┐
│              Federated Networks                       │
│                                                       │
│  ┌─────────────┐      ┌─────────────┐               │
│  │   grewal    │◄────►│    alice    │               │
│  │   Network   │      │   Network   │               │
│  └──────┬──────┘      └──────┬──────┘               │
│         │                    │                       │
│         └────────┬───────────┘                       │
│                  │                                   │
│                  ▼                                   │
│         ┌─────────────┐                              │
│         │     bob     │                              │
│         │   Network   │                              │
│         └─────────────┘                              │
└──────────────────────────────────────────────────────┘

Routing Table (grewal relay):
- laptop.grewal     → direct (own network)
- desktop.alice     → peer via alice relay
- server.bob        → peer via bob relay
- phone.alice       → peer via alice relay

Cross-Network Connection:
laptop.grewal → server.bob

1. laptop connects to grewal relay
2. grewal relay forwards to bob relay (federation)
3. bob relay connects to server.bob
4. Connection established
```

---

## Mobile Architecture

### iOS (NEPacketTunnelProvider)

```
┌─────────────────────────────────────────────────────┐
│  iOS Application Layer                              │
│  ┌────────┐  ┌──────────┐  ┌──────────────┐       │
│  │  SSH   │  │  Safari  │  │  Plex App    │       │
│  └───┬────┘  └────┬─────┘  └──────┬───────┘       │
└──────┼────────────┼────────────────┼───────────────┘
       └────────────┴────────────────┘
                    │
              IP packets to
              10.64.x.x
                    │
                    ▼
       ┌────────────────────────────┐
       │   iOS Network Stack        │
       └────────────┬───────────────┘
                    │
                    ▼
       ┌────────────────────────────┐
       │  NEPacketTunnelProvider    │
       │  (peer-up VPN extension)   │
       │                            │
       │  1. Capture packets        │
       │  2. Extract dest IP        │
       │  3. Map to peer ID         │
       │  4. Route via P2P          │
       └────────────┬───────────────┘
                    │
                    ▼
       ┌────────────────────────────┐
       │   libp2p-go (gomobile)     │
       │   P2P networking           │
       └────────────────────────────┘
```

### Android (VPNService)

Similar to iOS but with full VPNService API access:
- Create TUN interface
- Route all 10.64.0.0/16 traffic through app
- Full libp2p-go integration (easier than iOS)

---

## Performance Considerations

### Connection Optimization

1. **Relay vs Direct**:
   - Always attempt DCUtR for direct connection
   - Fall back to relay if hole-punching fails
   - Monitor connection quality and retry DCUtR periodically

2. **Connection Pooling**:
   - Reuse P2P streams for multiple requests
   - Multiplex services over single connection
   - Keep-alive mechanisms

3. **Bandwidth Management**:
   - QoS for different service types
   - Rate limiting per service
   - Bandwidth monitoring and alerts

### Caching

- DNS responses cached locally (TTL: 5 minutes)
- Peer ID → multiaddr mapping cached
- Federation routing table cached with periodic refresh

---

## Security Considerations

### Threat Model

**Threats Addressed**:
- ✅ Unauthorized peer access (ConnectionGater)
- ✅ Man-in-the-middle (libp2p Noise encryption)
- ✅ Replay attacks (Noise protocol nonces)
- ✅ Relay bandwidth theft (relay authentication)

**Threats NOT Addressed** (out of scope):
- ❌ Relay compromise (relay can see metadata, not content)
- ❌ Peer key compromise (users must secure private keys)
- ❌ DoS attacks (rate limiting planned for future)

### Best Practices

1. **Key Management**:
   - Private keys: 0600 permissions
   - authorized_keys: 0600 permissions
   - Never commit keys to git

2. **Network Segmentation**:
   - Use per-service authorized_keys when needed
   - Limit service exposure (disable unused services)
   - Audit authorized_keys regularly

3. **Relay Security**:
   - Enable relay authentication in production
   - Monitor relay bandwidth usage
   - Use non-standard ports

---

## Scalability

### Current Limitations

- **Relay bandwidth**: Limited by VPS plan (~1TB/month)
- **Connections per relay**: Limited by file descriptors (~1000-10000)
- **DHT lookups**: Slow for large networks (10-30 seconds)

### Future Improvements

- Multiple relay failover/load balancing
- Relay-to-relay mesh for redundancy
- Optimized peer routing (shortest path)
- Distributed hash table optimization
- Connection multiplexing

---

## Technology Stack

**Core**:
- Go 1.25+
- libp2p v0.38.2+ (networking)
- Kademlia DHT (peer discovery)
- Noise protocol (encryption)
- QUIC transport (performance)

**Optional**:
- Ethereum (blockchain naming)
- IPFS (distributed storage)
- gomobile (iOS/Android)

---

**Last Updated**: 2026-02-13
**Architecture Version**: 2.0 (Phase 4)
