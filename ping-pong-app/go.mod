module ping-pong-app

go 1.23

require (
     github.com/libp2p/go-libp2p v0.37.4
     github.com/libp2p/go-libp2p-kad-dht v0.26.1
     golang.org/x/net v0.33.0 // for IPv6 support (e.g., dialing ws+tcp)
)

require (
     // indirect deps (go mod tidy will fetch them)
     github.com/libp2p/go-libp2p-pipe v0.0.0 // not needed but avoids warnings
)
