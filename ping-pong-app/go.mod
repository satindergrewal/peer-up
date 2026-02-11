mkdir ping-pong-app && cd ping-pong-app
go mod init github.com/yourname/ping-pong-app

# Core libp2p deps + relay + websocket
go get github.com/libp2p/go-libp2p@v0.34.0 \
     github.com/libp2p/go-libp2p-pnet@v0.2.0 \
     github.com/libp2p/go-libp2p-peerstore@v0.5.0 \
     github.com/libp2p/go-libp2p-transport-upgrader@v0.6.0 \
     github.com/libp2p/go-libp2p-circuit-v2@v0.3.0 \
     github.com/libp2p/go-libp2p-quic-transport@v1.14.0 \
     github.com/libp2p/go-libp2p-ws-transport@v1.0.0 \
     github.com/libp2p/go-libp2p-tls@v1.0.0 \
     github.com/libp2p/go-libp2p-mplex@v0.5.0 \
     github.com/multiformats/go-multiaddr@v1.6.0 \
     github.com/multiformats/go-multiaddr-net@v1.3.1
