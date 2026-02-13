# peer-up: Decentralized P2P Network Infrastructure

A libp2p-based peer-to-peer network platform that enables secure connections across CGNAT networks with SSH-style authentication, service exposure, and local-first naming.

## Vision

**peer-up** is evolving from a simple NAT traversal tool into a comprehensive **decentralized P2P network infrastructure** that:

- Connects your devices across CGNAT/firewall barriers (Starlink, mobile networks)
- Exposes local services (SSH, XRDP, HTTP, SMB, custom protocols) through P2P connections
- Federates networks - Connect your network to friends' networks
- Works on mobile - iOS/Android apps with VPN-like functionality
- Flexible naming - Local names, network-scoped domains, optional blockchain anchoring
- Reusable library - Import `pkg/p2pnet` in your own Go projects

## Current Status (Phase 4A Complete)

- **Configuration-Based** - YAML config files, no hardcoded values
- **SSH-Style Authentication** - `authorized_keys` file for peer access control
- **NAT Traversal** - Works through Starlink CGNAT using relay + hole-punching
- **Persistent Identity** - Ed25519 keypairs saved to files
- **DHT Discovery** - Find peers using rendezvous on Kademlia DHT
- **Direct Connection Upgrade** - DCUtR attempts hole-punching for direct P2P
- **Key Management Tool** - `keytool` CLI for managing keypairs and authorized_keys
- **Service Exposure** - Expose any TCP service (SSH, XRDP, HTTP, etc.) via P2P
- **Reusable Library** - `pkg/p2pnet` package for building P2P applications
- **Unified Client** - Single `peerup` binary with subcommands (proxy, ping)
- **Name Resolution** - Map friendly names to peer IDs in config

## The Problem

Starlink uses Carrier-Grade NAT (CGNAT) on IPv4, and blocks inbound IPv6 connections via router firewall. This makes direct peer-to-peer connections impossible without a relay.

## The Solution

```
┌──────────┐         ┌──────────────┐         ┌──────────────┐
│  Client   │───────▶│ Relay Server │◀────────│  Home Node   │
│  (Phone)  │ outbound    (VPS)   outbound   │  (Linux/Mac) │
└──────────┘         └──────────────┘         └──────────────┘
                           │
                     Both connect OUTBOUND
                     Authentication enforced
```

1. **Relay Server** (VPS) - Circuit relay with optional authentication via `authorized_keys`
2. **Home Node** - Exposes local services, accepts only authorized peers
3. **Client (`peerup`)** - Connects to home node services through the relay

## Project Structure

```
├── cmd/
│   ├── home-node/              # Home node binary
│   │   └── main.go
│   ├── peerup/                 # Client binary (proxy + ping subcommands)
│   │   ├── main.go
│   │   ├── cmd_proxy.go
│   │   └── cmd_ping.go
│   └── keytool/                # Key management CLI
│       ├── main.go
│       └── commands/
├── pkg/p2pnet/                 # Reusable P2P networking library
│   ├── network.go              # Core network setup, relay helpers, name resolution
│   ├── service.go              # Service registry and management
│   ├── proxy.go                # Bidirectional TCP↔Stream proxy with half-close
│   ├── naming.go               # Local name resolution (name → peer ID)
│   └── identity.go             # Ed25519 identity management
├── internal/
│   ├── config/                 # YAML configuration loading
│   │   ├── config.go
│   │   └── loader.go
│   └── auth/                   # Authentication system
│       ├── authorized_keys.go
│       └── gater.go
├── relay-server/               # VPS relay node (separate module)
│   ├── main.go
│   └── relay-server.service
├── configs/                    # Sample configuration files
│   ├── home-node.sample.yaml
│   ├── client-node.sample.yaml
│   ├── relay-server.sample.yaml
│   └── authorized_keys.sample
├── go.mod                      # Single root module
└── ROADMAP.md
```

## Quick Start

### 1. Deploy Relay Server (VPS)

```bash
cd relay-server

# Create config from sample
cp ../configs/relay-server.sample.yaml relay-server.yaml
# Edit relay-server.yaml if needed (defaults are fine)

# Build and run
go build -o relay-server
./relay-server
```

Copy the **Relay Peer ID** from the output - you'll need it for the next steps.

### 2. Set Up Home Node

The home node runs on your home computer and exposes local services.

```bash
# Build from repo root (outputs binary to current directory)
go build -o home-node ./cmd/home-node

# Create a working directory for the home node
mkdir -p ~/home-node && cd ~/home-node

# Create config from sample
cp /path/to/peer-up/configs/home-node.sample.yaml home-node.yaml

# Edit home-node.yaml:
# 1. Set relay address and peer ID from step 1
# 2. Enable services you want to expose (ssh, xrdp, web, etc.)

# Create authorized_keys file
cp /path/to/peer-up/configs/authorized_keys.sample authorized_keys
# Add client peer IDs (one per line)

# Run
./home-node
```

Copy the **Home Node Peer ID** from the output.

### 3. Set Up Client (peerup)

The `peerup` client runs on your phone/laptop and connects to home node services.

```bash
# Build from repo root
go build -o peerup ./cmd/peerup

# Create a working directory for the client
mkdir -p ~/client-node && cd ~/client-node

# Create config from sample
cp /path/to/peer-up/configs/client-node.sample.yaml client-node.yaml

# Edit client-node.yaml:
# 1. Set relay address and peer ID from step 1
# 2. Add name mappings (optional, for convenience)
# 3. Set key_file for persistent identity (recommended)

# Create authorized_keys file
cp /path/to/peer-up/configs/authorized_keys.sample authorized_keys
# Add home node peer ID
```

### 4. Use Services

**Important:** Always run `peerup` from the directory containing `client-node.yaml`.

```bash
cd ~/client-node

# SSH to home node
../peerup proxy home ssh 2222
# In another terminal: ssh -p 2222 user@localhost

# Remote desktop (XRDP)
../peerup proxy home xrdp 13389
# In another terminal: xfreerdp /v:localhost:13389 /u:user

# Any TCP service
../peerup proxy home web 8080
# In browser: http://localhost:8080

# Ping home node (test connectivity)
../peerup ping home
```

You can use either a name (like `home`) or the full peer ID:
```bash
../peerup proxy 12D3KooWLutPZ... ssh 2222
```

## Building

All binaries are built from the repo root using the single Go module:

```bash
# Build home node
go build -o home-node ./cmd/home-node

# Build client (peerup)
go build -o peerup ./cmd/peerup

# Build keytool
go build -o keytool ./cmd/keytool

# Build relay server (separate module)
cd relay-server && go build -o relay-server
```

**Cross-compile for Linux** (e.g., to deploy home-node on a Linux server):
```bash
GOOS=linux GOARCH=amd64 go build -o home-node ./cmd/home-node
```

## Configuration

### Home Node Config (`home-node.yaml`)

```yaml
identity:
  key_file: "home_node.key"

network:
  listen_addresses:
    - "/ip4/0.0.0.0/tcp/9100"
    - "/ip4/0.0.0.0/udp/9100/quic-v1"
  force_private_reachability: true  # Required for CGNAT

relay:
  addresses:
    - "/ip4/YOUR_VPS_IP/tcp/7777/p2p/YOUR_RELAY_PEER_ID"
  reservation_interval: "2m"

discovery:
  rendezvous: "my-private-p2p-network"

security:
  authorized_keys_file: "authorized_keys"
  enable_connection_gating: true

protocols:
  ping_pong:
    enabled: true
    id: "/pingpong/1.0.0"

# Expose local services through P2P
services:
  ssh:
    enabled: true
    local_address: "localhost:22"
  xrdp:
    enabled: true
    local_address: "localhost:3389"
  web:
    enabled: false
    local_address: "localhost:80"
  plex:
    enabled: false
    local_address: "localhost:32400"
```

### Client Node Config (`client-node.yaml`)

```yaml
identity:
  key_file: "client_node.key"

network:
  listen_addresses:
    - "/ip4/0.0.0.0/tcp/0"
    - "/ip4/0.0.0.0/udp/0/quic-v1"

relay:
  addresses:
    - "/ip4/YOUR_VPS_IP/tcp/7777/p2p/YOUR_RELAY_PEER_ID"

discovery:
  rendezvous: "my-private-p2p-network"

security:
  authorized_keys_file: "authorized_keys"
  enable_connection_gating: true

protocols:
  ping_pong:
    enabled: true
    id: "/pingpong/1.0.0"

# Map friendly names to peer IDs
names:
  home: "YOUR_HOME_NODE_PEER_ID"
```

### Authorized Keys Format

```bash
# File: authorized_keys
# Format: <peer_id> # optional comment
12D3KooWLCavCP1Pma9NGJQnGDQhgwSjgQgupWprZJH4w1P3HCVL  # my-laptop
12D3KooWPjceQrSwdWXPyLLeABRXmuqt69Rg3sBYUXCUVwbj7QbA  # my-phone
```

## peerup Commands

### `peerup proxy` - Forward TCP to remote service

```
Usage: peerup proxy <target> <service> <local-port>

Arguments:
  target       Peer ID or name from config (e.g., "home")
  service      Service name as defined in home node config (e.g., "ssh", "xrdp")
  local-port   Local TCP port to listen on

Examples:
  peerup proxy home ssh 2222          # SSH via name
  peerup proxy home xrdp 13389       # Remote desktop
  peerup proxy 12D3KooW... ssh 2222  # SSH via peer ID
```

### `peerup ping` - Test connectivity

```
Usage: peerup ping <target>

Arguments:
  target    Peer ID or name from config

Examples:
  peerup ping home
  peerup ping 12D3KooW...
```

## Library (`pkg/p2pnet`)

The `pkg/p2pnet` package can be imported into your own Go projects:

```go
import "github.com/satindergrewal/peer-up/pkg/p2pnet"

// Create a P2P network
net, _ := p2pnet.New(&p2pnet.Config{
    KeyFile:      "myapp.key",
    EnableRelay:  true,
    RelayAddrs:   []string{"/ip4/.../tcp/7777/p2p/..."},
})

// Expose a local service
net.ExposeService("api", "localhost:8080")

// Connect to a peer's service
conn, _ := net.ConnectToService(peerID, "api")

// Name resolution
net.LoadNames(map[string]string{"home": "12D3KooW..."})
peerID, _ := net.ResolveName("home")

// Add relay addresses for a remote peer
net.AddRelayAddressesForPeer(relayAddrs, peerID)

// Create a TCP listener that proxies to a remote service
listener, _ := p2pnet.NewTCPListener("localhost:8080", func() (p2pnet.ServiceConn, error) {
    return net.ConnectToService(peerID, "api")
})
listener.Serve()
```

## Authentication System

Two layers of defense:

1. **ConnectionGater** (network level) - Blocks unauthorized peers during connection handshake
2. **Protocol handler validation** (application level) - Double-checks authorization before processing requests

### How It Works

1. **Home node** loads `authorized_keys` at startup
2. When a peer attempts to connect, `InterceptSecured()` checks the peer ID
3. If not authorized, connection is **DENIED** at network level
4. If authorized, connection is allowed and protocol handler performs secondary check

### Fail-Safe Defaults

- If `enable_connection_gating: true` but no `authorized_keys` file: **refuses to start**
- If `authorized_keys` is empty: **warns loudly** but allows (for initial setup)
- Home node: **allows all outbound** connections (for DHT, relay, etc.)
- Home node: **blocks all unauthorized inbound** connections

## Security Notes

### File Permissions

```bash
chmod 600 *.key              # Private keys: owner read/write only
chmod 600 authorized_keys    # SSH-style: owner read/write only
chmod 644 *.yaml             # Configs: readable by all
```

### Relay Server Security

The relay server supports `authorized_keys` to restrict who can make reservations. Enable authentication in production to protect your VPS bandwidth.

## Architecture

### Relay Circuit (Circuit Relay v2)

1. Home node connects outbound to relay and makes a **reservation**
2. Client connects outbound to relay
3. Client dials home via circuit address:
   ```
   /ip4/<RELAY_IP>/tcp/7777/p2p/<RELAY_PEER_ID>/p2p-circuit/p2p/<HOME_PEER_ID>
   ```
4. Relay bridges the connection - both sides only made outbound connections

### Hole-Punching (DCUtR)

After relay connection is established, libp2p attempts **Direct Connection Upgrade through Relay**:
- If successful: subsequent data flows directly (no relay bandwidth)
- If failed (symmetric NAT): continues using relay

### Peer Discovery (Kademlia DHT)

Home node **advertises** on DHT using rendezvous string.
Client node **searches** DHT for the rendezvous string to find home node's peer ID and addresses.

### Bidirectional Proxy

The TCP proxy uses the half-close pattern (inspired by Go stdlib's `httputil.ReverseProxy`):
- When one direction finishes sending, it signals `CloseWrite` instead of closing the connection
- The other direction can continue sending until it also finishes
- This prevents premature connection closure and works correctly with protocols like SSH and XRDP

## keytool - Key Management Utility

### Building

```bash
cd cmd/keytool
go build -o keytool
```

### Commands

```bash
# Generate new Ed25519 keypair
keytool generate my-node.key

# Extract peer ID from key file
keytool peerid home_node.key

# Validate authorized_keys file
keytool validate authorized_keys

# Add peer to authorized_keys
keytool authorize 12D3KooW... --comment "laptop" --file authorized_keys

# Remove peer from authorized_keys
keytool revoke 12D3KooW... --file authorized_keys
```

### Common Workflows

```bash
# Initial setup
keytool generate home-node.key
keytool generate client-node.key
keytool peerid client-node.key
keytool authorize <CLIENT_PEER_ID> --comment "my-phone" --file authorized_keys
```

## Running as a Service (systemd)

### Relay Server

```bash
sudo cp relay-server/relay-server.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable relay-server
sudo systemctl start relay-server
```

### Home Node

Create `/etc/systemd/system/home-node.service`:
```ini
[Unit]
Description=peer-up Home Node
After=network-online.target
Wants=network-online.target

[Service]
ExecStart=/path/to/home-node
WorkingDirectory=/path/to/home-node-config-dir
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

## Troubleshooting

| Issue | Solution |
|-------|----------|
| `Failed to load config` | Ensure you're running from the directory containing the YAML config |
| `Cannot resolve target` | Add name mapping to `names:` section in `client-node.yaml` |
| `DENIED inbound connection` | Add peer ID to `authorized_keys` and restart home-node |
| Home node shows no `/p2p-circuit` addresses | Check `force_private_reachability: true` and relay address |
| `protocols not supported: [/libp2p/circuit/relay/0.2.0/hop]` | Relay service not running |
| XRDP window manager crashes | Ensure no conflicting physical desktop session for the same user |
| `failed to sufficiently increase receive buffer size` | Warning only: `sudo sysctl -w net.core.rmem_max=7500000` |

## Bandwidth Considerations

- **Relay-based connection**: Limited by relay VPS bandwidth (~1TB/month on $5 Linode)
- **After DCUtR upgrade**: Direct P2P connection, no relay bandwidth used
- **Starlink symmetric NAT**: DCUtR often fails, relay remains in use

## Roadmap

See [ROADMAP.md](ROADMAP.md) for detailed multi-phase implementation plan.

### Phase 1-3: Foundation - COMPLETE
- Configuration system (YAML-based)
- SSH-style authentication (ConnectionGater + authorized_keys)
- Relay-based NAT traversal
- `keytool` CLI utility for key management

### Phase 4A: Core Library & Service Registry - COMPLETE
- `pkg/p2pnet` reusable library
- Service registry and bidirectional TCP proxy
- `cmd/` layout with single Go module
- `peerup` unified client with subcommands
- Local name resolution
- Tested: SSH, XRDP, generic TCP across LAN and 5G

### Phase 4B: Desktop Gateway - Next
- Multi-mode daemon: SOCKS / DNS / TUN
- Virtual network overlay
- Local DNS server (`.p2p` TLD)

### Phase 4C-4E: Federation, Mobile, Advanced Naming
See [ROADMAP.md](ROADMAP.md).

## Dependencies

- [go-libp2p](https://github.com/libp2p/go-libp2p) v0.47.0
- [go-libp2p-kad-dht](https://github.com/libp2p/go-libp2p-kad-dht) v0.28.1
- [go-multiaddr](https://github.com/multiformats/go-multiaddr)
- [gopkg.in/yaml.v3](https://gopkg.in/yaml.v3) v3.0.1
- [urfave/cli](https://github.com/urfave/cli) v1.22.17 (keytool)
- [fatih/color](https://github.com/fatih/color) v1.18.0 (keytool)

## License

MIT

## Contributing

This is a personal project, but issues and PRs are welcome!

**Testing checklist for PRs:**
- [ ] `go build ./cmd/home-node` succeeds
- [ ] `go build ./cmd/peerup` succeeds
- [ ] `go build ./cmd/keytool` succeeds
- [ ] Config files load without errors
- [ ] Unauthorized peer is denied
- [ ] Authorized peer connects successfully
- [ ] Service proxy works (SSH, XRDP, or other TCP)

---

**Built with libp2p** - Peer-to-peer networking that just works.
