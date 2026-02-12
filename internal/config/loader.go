package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// LoadHomeNodeConfig loads home node configuration from a YAML file
func LoadHomeNodeConfig(path string) (*HomeNodeConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
	}

	// Parse YAML with custom unmarshaling for durations
	var rawConfig struct {
		Identity  IdentityConfig  `yaml:"identity"`
		Network   NetworkConfig   `yaml:"network"`
		Relay     struct {
			Addresses           []string `yaml:"addresses"`
			ReservationInterval string   `yaml:"reservation_interval"`
		} `yaml:"relay"`
		Discovery DiscoveryConfig `yaml:"discovery"`
		Security  SecurityConfig  `yaml:"security"`
		Protocols ProtocolsConfig `yaml:"protocols"`
	}

	if err := yaml.Unmarshal(data, &rawConfig); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Parse duration
	reservationInterval, err := time.ParseDuration(rawConfig.Relay.ReservationInterval)
	if err != nil {
		return nil, fmt.Errorf("invalid reservation_interval: %w", err)
	}

	config := &HomeNodeConfig{
		Identity:  rawConfig.Identity,
		Network:   rawConfig.Network,
		Discovery: rawConfig.Discovery,
		Security:  rawConfig.Security,
		Protocols: rawConfig.Protocols,
		Relay: RelayConfig{
			Addresses:           rawConfig.Relay.Addresses,
			ReservationInterval: reservationInterval,
		},
	}

	return config, nil
}

// LoadClientNodeConfig loads client node configuration from a YAML file
func LoadClientNodeConfig(path string) (*ClientNodeConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
	}

	// Parse YAML with custom unmarshaling for durations
	var rawConfig struct {
		Identity  IdentityConfig  `yaml:"identity"`
		Network   NetworkConfig   `yaml:"network"`
		Relay     struct {
			Addresses           []string `yaml:"addresses"`
			ReservationInterval string   `yaml:"reservation_interval"`
		} `yaml:"relay"`
		Discovery DiscoveryConfig `yaml:"discovery"`
		Security  SecurityConfig  `yaml:"security"`
		Protocols ProtocolsConfig `yaml:"protocols"`
	}

	if err := yaml.Unmarshal(data, &rawConfig); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Parse duration
	reservationInterval, err := time.ParseDuration(rawConfig.Relay.ReservationInterval)
	if err != nil {
		return nil, fmt.Errorf("invalid reservation_interval: %w", err)
	}

	config := &ClientNodeConfig{
		Identity:  rawConfig.Identity,
		Network:   rawConfig.Network,
		Discovery: rawConfig.Discovery,
		Security:  rawConfig.Security,
		Protocols: rawConfig.Protocols,
		Relay: RelayConfig{
			Addresses:           rawConfig.Relay.Addresses,
			ReservationInterval: reservationInterval,
		},
	}

	return config, nil
}

// LoadRelayServerConfig loads relay server configuration from a YAML file
func LoadRelayServerConfig(path string) (*RelayServerConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
	}

	var config RelayServerConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return &config, nil
}

// ValidateHomeNodeConfig validates home node configuration
func ValidateHomeNodeConfig(cfg *HomeNodeConfig) error {
	if cfg.Identity.KeyFile == "" {
		return fmt.Errorf("identity.key_file is required")
	}
	if len(cfg.Network.ListenAddresses) == 0 {
		return fmt.Errorf("network.listen_addresses must contain at least one address")
	}
	if len(cfg.Relay.Addresses) == 0 {
		return fmt.Errorf("relay.addresses must contain at least one address")
	}
	if cfg.Discovery.Rendezvous == "" {
		return fmt.Errorf("discovery.rendezvous is required")
	}
	if cfg.Protocols.PingPong.ID == "" {
		return fmt.Errorf("protocols.ping_pong.id is required")
	}
	if cfg.Security.EnableConnectionGating && cfg.Security.AuthorizedKeysFile == "" {
		return fmt.Errorf("security.authorized_keys_file is required when connection gating is enabled")
	}
	return nil
}

// ValidateClientNodeConfig validates client node configuration
func ValidateClientNodeConfig(cfg *ClientNodeConfig) error {
	if len(cfg.Network.ListenAddresses) == 0 {
		return fmt.Errorf("network.listen_addresses must contain at least one address")
	}
	if len(cfg.Relay.Addresses) == 0 {
		return fmt.Errorf("relay.addresses must contain at least one address")
	}
	if cfg.Discovery.Rendezvous == "" {
		return fmt.Errorf("discovery.rendezvous is required")
	}
	if cfg.Protocols.PingPong.ID == "" {
		return fmt.Errorf("protocols.ping_pong.id is required")
	}
	if cfg.Security.EnableConnectionGating && cfg.Security.AuthorizedKeysFile == "" {
		return fmt.Errorf("security.authorized_keys_file is required when connection gating is enabled")
	}
	return nil
}

// ValidateRelayServerConfig validates relay server configuration
func ValidateRelayServerConfig(cfg *RelayServerConfig) error {
	if cfg.Identity.KeyFile == "" {
		return fmt.Errorf("identity.key_file is required")
	}
	if len(cfg.Network.ListenAddresses) == 0 {
		return fmt.Errorf("network.listen_addresses must contain at least one address")
	}
	if cfg.Security.EnableConnectionGating && cfg.Security.AuthorizedKeysFile == "" {
		return fmt.Errorf("security.authorized_keys_file is required when connection gating is enabled")
	}
	return nil
}
