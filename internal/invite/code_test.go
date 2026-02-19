package invite

import (
	"strings"
	"testing"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
)

func TestEncodeDecodeRoundTrip(t *testing.T) {
	priv, _, _ := crypto.GenerateKeyPair(crypto.Ed25519, -1)
	pid, _ := peer.IDFromPrivateKey(priv)

	token, err := GenerateToken()
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}

	data := &InviteData{
		Token:     token,
		RelayAddr: "/ip4/203.0.113.50/tcp/7777/p2p/12D3KooWRzaGMTqQbRHNMZkAYj8ALUXoK99qSjhiFLanDoVWK9An",
		PeerID:    pid,
	}

	code, err := Encode(data)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	t.Logf("Invite code (%d chars): %s", len(code), code)

	decoded, err := Decode(code)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}

	if token != decoded.Token {
		t.Errorf("Token mismatch")
	}
	if data.RelayAddr != decoded.RelayAddr {
		t.Errorf("RelayAddr mismatch: got %s, want %s", decoded.RelayAddr, data.RelayAddr)
	}
	if data.PeerID != decoded.PeerID {
		t.Errorf("PeerID mismatch: got %s, want %s", decoded.PeerID, data.PeerID)
	}
}

func TestTokenHex(t *testing.T) {
	token, err := GenerateToken()
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}
	hex := TokenHex(token)
	if len(hex) != 16 { // 8 bytes = 16 hex chars
		t.Errorf("TokenHex length = %d, want 16", len(hex))
	}
	// Should be valid hex
	for _, c := range hex {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Errorf("TokenHex contains non-hex char: %c", c)
		}
	}
}

func TestGenerateTokenUniqueness(t *testing.T) {
	t1, _ := GenerateToken()
	t2, _ := GenerateToken()
	if t1 == t2 {
		t.Error("two tokens should be different")
	}
}

func TestDecodeInvalid(t *testing.T) {
	_, err := Decode("not-a-valid-code")
	if err == nil {
		t.Error("expected error for invalid code")
	}

	_, err = Decode("")
	if err == nil {
		t.Error("expected error for empty code")
	}
}

// TestDecodeRejectsTrailingJunk verifies that Decode rejects invite codes
// with trailing data appended (e.g., from "peerup join CODE --name laptop"
// where Go's flag package concatenates --name and laptop into the code).
func TestDecodeRejectsTrailingJunk(t *testing.T) {
	priv, _, _ := crypto.GenerateKeyPair(crypto.Ed25519, -1)
	pid, _ := peer.IDFromPrivateKey(priv)

	token, err := GenerateToken()
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}

	data := &InviteData{
		Token:     token,
		RelayAddr: "/ip4/203.0.113.50/tcp/7777/p2p/12D3KooWQvzCBP1MdU6g3UC6rUwHtDkbMUWQKDapmHqQFPqZqTn7",
		PeerID:    pid,
	}

	code, err := Encode(data)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	// Simulate: fs.Args() = [code, "--name", "laptop"]
	// strings.Join(fs.Args(), "") = code + "--name" + "laptop"
	corrupted := strings.Join([]string{code, "--name", "laptop"}, "")

	_, err = Decode(corrupted)
	if err == nil {
		t.Error("Decode should reject invite code with trailing junk from --name flag")
	} else {
		t.Logf("Correctly rejected: %v", err)
	}

	// Also test with just random base32 junk appended
	junk := code + "AAAA"
	_, err = Decode(junk)
	if err == nil {
		t.Error("Decode should reject invite code with trailing base32 characters")
	} else {
		t.Logf("Correctly rejected junk: %v", err)
	}

	// Clean code should still work
	decoded, err := Decode(code)
	if err != nil {
		t.Fatalf("Clean code should decode: %v", err)
	}
	if decoded.PeerID != pid {
		t.Errorf("PeerID mismatch: got %s, want %s", decoded.PeerID, pid)
	}
}
