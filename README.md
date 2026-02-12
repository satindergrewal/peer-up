# peer-up: Secure P2P Connection for CGNAT Networks

A production-ready libp2p-based system for connecting devices across CGNAT networks (like Starlink) with SSH-style authentication and YAML configuration.

## Features

- ✅ **Configuration-Based** - YAML config files, no hardcoded values, no recompilation needed
- ✅ **SSH-Style Authentication** - `authorized_keys` file for peer access control
- ✅ **NAT Traversal** - Works through Starlink CGNAT using relay + hole-punching
- ✅ **Persistent Identity** - Ed25519 keypairs saved to files
- ✅ **DHT Discovery** - Find peers using rendezvous on Kademlia DHT
- ✅ **Direct Connection Upgrade** - DCUtR attempts hole-punching for direct P2P after relay connection

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

1. **Relay Server** (VPS) - Circuit relay with no authentication (initially)
2. **Home Node** - Accepts only authorized peers via `authorized_keys`
3. **Client Node** - Connects with persistent identity or ephemeral key

## Project Structure

```
├── configs/                     # Sample configuration files
│   ├── home-node.sample.yaml
│   ├── client-node.sample.yaml
│   ├── relay-server.sample.yaml
│   └── authorized_keys.sample
├── internal/                    # Shared packages
│   ├── config/                  # YAML configuration loading
│   │   ├── config.go
│   │   └── loader.go
│   └── auth/                    # Authentication system
│       ├── authorized_keys.go   # Parser for authorized_keys file
│       └── gater.go            # ConnectionGater implementation
├── relay-server/               # VPS relay node
│   ├── main.go
│   ├── go.mod
│   └── relay-server.service
├── home-node/                  # Home computer node
│   ├── main.go
│   └── go.mod
└── client-node/                # Phone/laptop client
    ├── main.go
    └── go.mod
```

## Quick Start

### 1. Deploy Relay Server (VPS)

```bash
# On your VPS
cd relay-server

# Create config from sample
cp ../configs/relay-server.sample.yaml relay-server.yaml

# Edit if needed (defaults are fine)
# Build and run
go build -o relay-server
./relay-server
```

**Copy the Relay Peer ID** from the output - you'll need it for the next steps.

### 2. Set Up Home Node

```bash
# On your home computer
cd home-node

# Create config from sample
cp ../configs/home-node.sample.yaml home-node.yaml

# Edit home-node.yaml:
# 1. Set relay address and peer ID from step 1
# 2. Set rendezvous string (keep default for now)
# 3. Enable/disable authentication

# Create authorized_keys file (if authentication enabled)
cp ../configs/authorized_keys.sample authorized_keys
# Add client peer IDs (one per line)

# Build and run
go build -o home-node
./home-node
```

**Copy the Home Node Peer ID** from the output.

### 3. Set Up Client Node

```bash
# On your phone/laptop
cd client-node

# Create config from sample
cp ../configs/client-node.sample.yaml client-node.yaml

# Edit client-node.yaml:
# 1. Set relay address and peer ID from step 1
# 2. Match rendezvous string with home-node
# 3. Set key_file for persistent identity (optional)

# Build and run
go build -o client-node
./client-node <HOME_PEER_ID>
```

## Configuration

### Home Node Config (`home-node.yaml`)

```yaml
identity:
  key_file: "home_node.key"  # Persistent peer identity

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
  rendezvous: "my-private-p2p-network"  # Custom rendezvous string
  bootstrap_peers: []  # Empty = use libp2p defaults

security:
  authorized_keys_file: "authorized_keys"
  enable_connection_gating: true  # Enforce authentication

protocols:
  ping_pong:
    enabled: true
    id: "/pingpong/1.0.0"
```

### Authorized Keys Format

```bash
# File: authorized_keys
# Format: <peer_id> # optional comment

12D3KooWLCavCP1Pma9NGJQnGDQhgwSjgQgupWprZJH4w1P3HCVL  # my-laptop
12D3KooWPjceQrSwdWXPyLLeABRXmuqt69Rg3sBYUXCUVwbj7QbA  # my-phone
```

## Authentication System

The authentication system uses two layers of defense:

1. **ConnectionGater** (network level) - Blocks unauthorized peers during connection handshake
2. **Protocol handler validation** (application level) - Double-checks authorization before processing requests

### How It Works

1. **Home node** loads `authorized_keys` at startup
2. When a peer attempts to connect, `InterceptSecured()` checks the peer ID
3. If not authorized → connection **DENIED** at network level
4. If authorized → connection allowed, protocol handler performs secondary check

### Testing Authentication

**Test 1: Unauthorized peer (should be denied)**
```bash
# Run client-node without adding its peer ID to authorized_keys
# Watch home-node logs:
# "DENIED inbound connection from unauthorized peer: 12D3KooW..."
```

**Test 2: Authorized peer (should work)**
```bash
# Add client peer ID to home-node/authorized_keys
# Restart home-node (to reload authorized_keys)
# Run client-node again
# Should connect successfully and receive pong
```

## Security Notes

### File Permissions

```bash
chmod 600 *.key              # Private keys: owner read/write only
chmod 600 authorized_keys    # SSH-style: owner read/write only
chmod 644 *.yaml            # Configs: readable by all
```

### Fail-Safe Defaults

- If `enable_connection_gating: true` but no `authorized_keys` file → **refuses to start**
- If `authorized_keys` is empty → **warns loudly** but allows (for initial setup)
- Home node: **Allows all outbound** connections (for DHT, relay, etc.)
- Home node: **Blocks all unauthorized inbound** connections

### Relay Server Security

**Initial implementation:** Relay is open (no authentication) to reduce friction.

**Future enhancement:** Add `authorized_keys` support to relay-server to restrict who can make reservations.

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
- If successful → subsequent data flows directly (no relay bandwidth)
- If failed (symmetric NAT) → continues using relay

### Peer Discovery (Kademlia DHT)

Home node **advertises** on DHT using rendezvous string.
Client node **searches** DHT for the rendezvous string to find home node's peer ID and addresses.

## Relay Server Details

The relay is minimal and private:

```go
h, err := libp2p.New(
    libp2p.Identity(priv),
    libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/7777"),
)
relayv2.New(h, relayv2.WithInfiniteLimits())
```

**Key design decisions:**
- **No DHT participation** - Avoids IPFS swarm traffic
- **Non-standard port (7777)** - Further avoids IPFS discovery
- **`WithInfiniteLimits()`** - Safe for private relay with known peers
- **Manual `relayv2.New()`** - More reliable than `EnableRelayService()` option

### Running as a Service (systemd)

```bash
# On VPS
sudo cp relay-server.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable relay-server
sudo systemctl start relay-server
sudo systemctl status relay-server
```

## Troubleshooting

| Issue | Solution |
|-------|----------|
| `Failed to load config` | Create config files from samples in `configs/` |
| `Invalid configuration` | Check YAML syntax, required fields |
| `Connection gating enabled but no authorized_keys_file` | Specify path in config |
| `DENIED inbound connection` | Add peer ID to `authorized_keys` and restart home-node |
| `authorized_keys file is empty` | Add authorized peer IDs (one per line) |
| Home node shows no `/p2p-circuit` addresses | Check `force_private_reachability: true` and relay address |
| `protocols not supported: [/libp2p/circuit/relay/0.2.0/hop]` | Relay service not running - check relay-server |
| Relay swarmed by peers | Don't enable DHT on relay, use non-standard port |
| `failed to sufficiently increase receive buffer size` | Warning only: `sudo sysctl -w net.core.rmem_max=7500000` |

## Key Lessons Learned

1. **`ForceReachabilityPrivate()`** is essential on home node behind CGNAT - without it, libp2p detects IPv6 and assumes it's reachable, skipping relay reservations

2. **ConnectionGater checks happen after crypto handshake** - peer ID is verified before authorization check (`InterceptSecured`)

3. **Configuration hot-reload not implemented** - Restart nodes after editing `authorized_keys` (file watcher planned for Phase 3)

4. **Relay should use `relayv2.New(host)` manually** - `libp2p.EnableRelayService()` as host option didn't reliably register hop protocol in testing

5. **Port 4001 = IPFS swarm traffic** - Use non-standard ports and skip DHT on relay to avoid hundreds of IPFS peer connections

## Bandwidth Considerations

- **Relay-based connection**: Limited by relay VPS bandwidth (~1TB/month on $5 Linode)
- **After DCUtR upgrade**: Direct P2P connection, no relay bandwidth used
- **Starlink symmetric NAT**: DCUtR often fails, relay remains in use
- **Workaround**: Place Starlink router in bypass mode + custom router with IPv6 firewall configuration

## Future Enhancements (Phase 3+)

### Phase 3: Enhanced Usability
- [ ] `keytool` utility for key management
  - Generate keypairs
  - Extract peer ID from key file
  - Validate authorized_keys file
  - Add/remove authorized peers
- [ ] Config validation CLI flag (`--validate-config`)
- [ ] Hot-reload for `authorized_keys` (using file watcher)

### Phase 4: Service Exposure
- [ ] SSH tunneling through P2P connection
- [ ] HTTP/HTTPS reverse proxy
- [ ] Per-service authorization (override global `authorized_keys`)
- [ ] Protocol naming: `/peerup/<service>/<version>`

### Phase 5: Mobile
- [ ] iOS app using `gomobile bind`
- [ ] Android app
- [ ] mDNS for local network discovery (bypass relay on home WiFi)

## Dependencies

- [go-libp2p](https://github.com/libp2p/go-libp2p) v0.38.2
- [go-libp2p-kad-dht](https://github.com/libp2p/go-libp2p-kad-dht) v0.28.1
- [go-multiaddr](https://github.com/multiformats/go-multiaddr) v0.14.0
- [gopkg.in/yaml.v3](https://gopkg.in/yaml.v3) v3.0.1

## License

MIT

## Contributing

This is a personal project, but issues and PRs are welcome!

**Testing checklist for PRs:**
- [ ] All three nodes build successfully
- [ ] Config files load without errors
- [ ] Unauthorized peer is denied
- [ ] Authorized peer connects successfully
- [ ] PING-PONG protocol works

---

**Built with libp2p** - Peer-to-peer networking that just works. 🚀
