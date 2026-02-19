package p2pnet

import (
	"strings"
	"testing"
)

func TestTruncateError(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"short", "connection refused", "connection refused"},
		{"multiline", "first line\nsecond line\nthird", "first line"},
		{"long", strings.Repeat("a", 250), strings.Repeat("a", 200) + "..."},
		{"multiline long first", strings.Repeat("x", 250) + "\nsecond", strings.Repeat("x", 200) + "..."},
		{"empty", "", ""},
		{"exactly 200", strings.Repeat("b", 200), strings.Repeat("b", 200)},
		{"newline at start", "\nsecond line", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncateError(tt.input)
			if got != tt.want {
				t.Errorf("truncateError(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseRelayAddrs(t *testing.T) {
	t.Run("valid single", func(t *testing.T) {
		addrs := []string{
			"/ip4/203.0.113.50/tcp/7777/p2p/12D3KooWRzaGMTqQbRHNMZkAYj8ALUXoK99qSjhiFLanDoVWK9An",
		}
		infos, err := ParseRelayAddrs(addrs)
		if err != nil {
			t.Fatalf("ParseRelayAddrs: %v", err)
		}
		if len(infos) != 1 {
			t.Fatalf("got %d infos, want 1", len(infos))
		}
		if infos[0].ID.String() != "12D3KooWRzaGMTqQbRHNMZkAYj8ALUXoK99qSjhiFLanDoVWK9An" {
			t.Errorf("peer ID = %s", infos[0].ID)
		}
	})

	t.Run("dedup same peer", func(t *testing.T) {
		addrs := []string{
			"/ip4/203.0.113.50/tcp/7777/p2p/12D3KooWRzaGMTqQbRHNMZkAYj8ALUXoK99qSjhiFLanDoVWK9An",
			"/ip4/203.0.113.50/udp/7778/quic-v1/p2p/12D3KooWRzaGMTqQbRHNMZkAYj8ALUXoK99qSjhiFLanDoVWK9An",
		}
		infos, err := ParseRelayAddrs(addrs)
		if err != nil {
			t.Fatalf("ParseRelayAddrs: %v", err)
		}
		if len(infos) != 1 {
			t.Fatalf("got %d infos, want 1 (dedup)", len(infos))
		}
		// Merged addresses
		if len(infos[0].Addrs) != 2 {
			t.Errorf("got %d addrs, want 2 (merged)", len(infos[0].Addrs))
		}
	})

	t.Run("empty list", func(t *testing.T) {
		infos, err := ParseRelayAddrs(nil)
		if err != nil {
			t.Fatalf("ParseRelayAddrs nil: %v", err)
		}
		if len(infos) != 0 {
			t.Errorf("got %d infos, want 0", len(infos))
		}
	})

	t.Run("invalid multiaddr", func(t *testing.T) {
		_, err := ParseRelayAddrs([]string{"not-a-multiaddr"})
		if err == nil {
			t.Error("expected error for invalid multiaddr")
		}
	})

	t.Run("missing peer ID", func(t *testing.T) {
		_, err := ParseRelayAddrs([]string{"/ip4/1.2.3.4/tcp/7777"})
		if err == nil {
			t.Error("expected error for addr without peer ID")
		}
	})
}
