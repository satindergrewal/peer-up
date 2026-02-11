package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
)

type Config struct {
	PeerID       string   `json:"peerID,omitempty"`
	PrivateKey   string   `json:"privateKey,omitempty"` // base64
	ListenAddrs  []string `json:"listenAddrs,omitempty"`
	RelayAddr    string   `json:"relayAddr,omitempty"`
	TargetPeerID string   `json:"targetPeerID,omitempty"`
	LocalPeerName string  `json:"localPeerName"`

	// Relay only
	EnableRelay     bool `json:"enableRelay,omitempty"`
	EnableRelayHop  bool `json:"enableRelayHop,omitempty"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	cfg := &Config{}
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	return cfg, nil
}

func SaveConfig(path string, cfg *Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

// GenerateKeyPair generates a new keypair and returns (privKey, pubKey)
func GenerateKeyPair() (crypto.PrivKey, crypto.PubKey, error) {
	return crypto.GenerateKeyPair(crypto.Ed25519, 2048)
}

// PeerIDFromPrivKey derives peer ID from private key
func PeerIDFromPrivKey(priv crypto.PrivKey) (peer.ID, error) {
	pub := priv.GetPublic()
	id, err := peer.IDFromPublicKey(pub)
	if err != nil {
		return "", err
	}
	return id, nil
}

// EncodePrivateKeyToBase64 returns base64 of raw private key bytes
func EncodePrivateKeyToBase64(priv crypto.PrivKey) (string, error) {
	b, err := crypto.MarshalPrivateKey(priv)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// DecodePrivateKeyFromBase64 decodes base64 to privKey
func DecodePrivateKeyFromBase64(s string) (crypto.PrivKey, error) {
	return crypto.UnmarshalPrivateKey([]byte(s))
}
