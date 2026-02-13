# Testing Guide: SSH Access via P2P Network

This guide walks through testing the complete peer-up system with SSH service exposure.

## Goal

Connect to your home computer's SSH server from a client device (laptop/phone) through the P2P network, traversing CGNAT/NAT using a relay server.

```
[Client Node] <--P2P--> [Relay Server] <--P2P--> [Home Node] --TCP--> [SSH Server :22]
   (Laptop)              (VPS/Cloud)           (Home PC behind CGNAT)
```

## Prerequisites

### 1. Three Machines/Terminals

- **Relay Server**: VPS with public IP (DigitalOcean, AWS, etc.)
- **Home Node**: Your home computer behind CGNAT/NAT
- **Client Node**: Laptop or another device

### 2. SSH Server Running

On your home computer:
```bash
# Check if SSH server is running
sudo systemctl status sshd  # or ssh on macOS

# Start if not running (Linux)
sudo systemctl start sshd

# macOS - enable in System Preferences > Sharing > Remote Login
```

### 3. Configurations Prepared

```bash
# Create config files from samples
cp configs/relay-server.sample.yaml relay-server/relay-server.yaml
cp configs/home-node.sample.yaml home-node/home-node.yaml
cp configs/client-node.sample.yaml client-node/client-node.yaml
```

---

## Step 1: Start Relay Server

### On VPS (Relay Server)

```bash
cd peer-up/relay-server

# Build
go build -o relay-server

# Run (it will create relay_server.key automatically)
./relay-server
```

**Expected output:**
```
=== Relay Server (Circuit Relay v2) ===
🆔 Relay Peer ID: 12D3KooWABC...XYZ
📍 Listening on:
  /ip4/YOUR_VPS_IP/tcp/7777
  /ip4/YOUR_VPS_IP/udp/7777/quic-v1
✅ Relay server is running!
```

**Save these values:**
- Relay Peer ID: `12D3KooWABC...XYZ`
- VPS IP: `YOUR_VPS_IP`

---

## Step 2: Configure Home Node

### Edit `home-node/home-node.yaml`

```yaml
identity:
  key_file: "home_node.key"

network:
  listen_addresses:
    - "/ip4/0.0.0.0/tcp/9100"
    - "/ip4/0.0.0.0/udp/9100/quic-v1"
  force_private_reachability: true  # CRITICAL for CGNAT

relay:
  # REPLACE with your relay server details from Step 1
  addresses:
    - "/ip4/YOUR_VPS_IP/tcp/7777/p2p/12D3KooWABC...XYZ"
  reservation_interval: "2m"

discovery:
  rendezvous: "my-private-p2p-network"
  bootstrap_peers: []

security:
  authorized_keys_file: "authorized_keys"
  enable_connection_gating: false  # Set true after adding client peer ID

protocols:
  ping_pong:
    enabled: true
    id: "/pingpong/1.0.0"

# IMPORTANT: Enable SSH service
services:
  ssh:
    enabled: true
    local_address: "localhost:22"
```

### Start Home Node

```bash
cd home-node

# Option 1: Use refactored version (recommended)
go build -o home-node-refactored main-refactored.go
./home-node-refactored

# Option 2: Use original version
go build -o home-node
./home-node
```

**Expected output:**
```
=== Home Node (Refactored with pkg/p2pnet) ===
Loaded configuration from home-node.yaml
Rendezvous: my-private-p2p-network

🏠 Peer ID: 12D3KooWHOME...ABC
✅ Connected to relay 12D3KooWABC...
Waiting for AutoRelay to establish reservations...
✅ Relay address: /ip4/YOUR_VPS_IP/tcp/7777/p2p/12D3KooWABC.../p2p-circuit/p2p/12D3KooWHOME...ABC
✅ Registered service: ssh (protocol: /peerup/ssh/1.0.0, local: localhost:22)
```

**Save the Home Node Peer ID**: `12D3KooWHOME...ABC`

---

## Step 3: Configure Client Node

### Edit `client-node/client-node.yaml`

```yaml
identity:
  key_file: ""  # Ephemeral identity (or set path for persistent)

network:
  listen_addresses:
    - "/ip4/0.0.0.0/tcp/0"  # Ephemeral port
    - "/ip4/0.0.0.0/udp/0/quic-v1"
  force_private_reachability: false

relay:
  # SAME relay as home-node
  addresses:
    - "/ip4/YOUR_VPS_IP/tcp/7777/p2p/12D3KooWABC...XYZ"
  reservation_interval: "2m"

discovery:
  rendezvous: "my-private-p2p-network"  # MUST MATCH home-node
  bootstrap_peers: []

security:
  authorized_keys_file: "authorized_keys"
  enable_connection_gating: false  # Disable for testing

protocols:
  ping_pong:
    enabled: true
    id: "/pingpong/1.0.0"

# Map "home" to your home-node peer ID
names:
  home: "12D3KooWHOME...ABC"  # From Step 2
```

### Start Client Node

```bash
cd client-node

go build -o client-node
./client-node
```

**Expected output:**
```
=== Client Node ===
📡 Searching for peers on rendezvous: my-private-p2p-network
✅ Found peer: 12D3KooWHOME...ABC
🔗 Connecting to peer via relay...
✅ Connected!
📤 Sending ping to 12D3KooWHOME...ABC
📥 Received: pong
```

---

## Step 4: Test SSH Connection via P2P

### Option A: Using Manual Script (Recommended)

Create `test-ssh-connection.go` in `client-node/`:

```go
package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/satindergrewal/peer-up/internal/config"
	"github.com/satindergrewal/peer-up/pkg/p2pnet"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <home-node-peer-id>", os.Args[0])
	}

	homeNodeID, err := peer.Decode(os.Args[1])
	if err != nil {
		log.Fatalf("Invalid peer ID: %v", err)
	}

	// Load config
	cfg, err := config.LoadClientNodeConfig("client-node.yaml")
	if err != nil {
		log.Fatalf("Config error: %v", err)
	}

	// Create P2P network with relay support
	net, err := p2pnet.New(&p2pnet.Config{
		KeyFile:            cfg.Identity.KeyFile,
		Config:             &config.Config{Network: cfg.Network},
		EnableRelay:        true,
		RelayAddrs:         cfg.Relay.Addresses,
		EnableNATPortMap:   true,
		EnableHolePunching: true,
	})
	if err != nil {
		log.Fatalf("P2P network error: %v", err)
	}
	defer net.Close()

	fmt.Printf("🆔 Client Peer ID: %s\n", net.PeerID())

	// Connect to home node via relay
	ctx := context.Background()
	h := net.Host()

	// Wait for relay connection
	fmt.Println("🔗 Connecting to home node via relay...")
	// Discovery would happen here normally, for now assume we know the peer

	// Open SSH service stream
	fmt.Println("🌐 Opening SSH service stream...")
	sshConn, err := net.ConnectToService(homeNodeID, "ssh")
	if err != nil {
		log.Fatalf("Failed to connect to SSH service: %v", err)
	}
	defer sshConn.Close()

	fmt.Println("✅ Connected to SSH service!")
	fmt.Println("📡 Creating local TCP proxy on localhost:2222")

	// Create local listener
	listener, err := net.Listen("tcp", "localhost:2222")
	if err != nil {
		log.Fatalf("Failed to create listener: %v", err)
	}
	defer listener.Close()

	fmt.Println("✅ SSH proxy ready!")
	fmt.Println("\n💡 Connect via SSH:")
	fmt.Println("   ssh -p 2222 username@localhost")
	fmt.Println("\nPress Ctrl+C to stop.")

	// Accept connections and forward to P2P
	for {
		localConn, err := listener.Accept()
		if err != nil {
			log.Printf("Accept error: %v", err)
			continue
		}

		go func(lc net.Conn) {
			defer lc.Close()

			// Open new stream for each connection
			sshStream, err := net.ConnectToService(homeNodeID, "ssh")
			if err != nil {
				log.Printf("Failed to open SSH stream: %v", err)
				return
			}
			defer sshStream.Close()

			// Bidirectional copy
			errCh := make(chan error, 2)
			go func() {
				_, err := io.Copy(sshStream, lc)
				errCh <- err
			}()
			go func() {
				_, err := io.Copy(lc, sshStream)
				errCh <- err
			}()

			<-errCh
		}(localConn)
	}
}
```

### Run the test:

```bash
cd client-node
go run test-ssh-connection.go 12D3KooWHOME...ABC
```

### Connect via SSH:

```bash
# In another terminal
ssh -p 2222 your_username@localhost
```

You should see your home computer's SSH prompt!

---

## Step 5: Verify End-to-End

### On Home Node - Check Logs

You should see:
```
📥 Incoming ssh connection from 12D3KooWCLIENT...
✅ Connected to local service localhost:22
```

### On Client Node

You should see successful SSH connection and be able to run commands on your home computer.

---

## Troubleshooting

### Relay Connection Failed

```
⚠️  Could not connect to relay
```

**Fix:**
- Verify VPS firewall allows TCP 7777 and UDP 7777
- Check relay server is actually running
- Verify relay peer ID is correct in configs

### No Relay Address

```
⚠️  No relay addresses yet
```

**Fix:**
- Ensure `force_private_reachability: true` in home-node config
- Wait 10-15 seconds for AutoRelay
- Check relay server logs for reservation requests

### SSH Service Not Found

```
Failed to connect to SSH service: protocol not supported
```

**Fix:**
- Verify `services.ssh.enabled: true` in home-node.yaml
- Check home-node logs for "Registered service: ssh"
- Ensure SSH protocol ID matches: `/peerup/ssh/1.0.0`

### Connection Refused on localhost:22

```
Failed to connect to local service localhost:22
```

**Fix:**
- Start SSH server on home computer
- Check: `sudo systemctl status sshd`
- Verify SSH is listening: `netstat -tlnp | grep :22`

### Discovery Not Working

```
📡 Searching for peers... (no results)
```

**Fix:**
- Verify both nodes use SAME `rendezvous` string
- Check DHT is bootstrapped (should see "Connected to X bootstrap peers")
- Wait 30-60 seconds for DHT propagation

---

## Next Steps

Once SSH works:

1. **Enable Security**: Add peer IDs to `authorized_keys` files
2. **Test Other Services**: Try HTTP, SMB, custom protocols
3. **Production Setup**: Use persistent identities, configure firewall rules
4. **Mobile App**: Build iOS/Android clients using the library

## Success Criteria

✅ Relay server running and accessible
✅ Home node gets relay address with `/p2p-circuit`
✅ Client finds home node via DHT
✅ SSH service stream opens successfully
✅ You can SSH into your home computer from anywhere

**You've now successfully traversed CGNAT/NAT using libp2p relay!** 🎉
