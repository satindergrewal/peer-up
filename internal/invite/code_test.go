package invite

import (
	"strings"
	"testing"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
)

func generateTestPeerID(t *testing.T) peer.ID {
	t.Helper()
	priv, _, _ := crypto.GenerateKeyPair(crypto.Ed25519, -1)
	pid, _ := peer.IDFromPrivateKey(priv)
	return pid
}

const testRelayAddr = "/ip4/203.0.113.50/tcp/7777/p2p/12D3KooWRzaGMTqQbRHNMZkAYj8ALUXoK99qSjhiFLanDoVWK9An"
const testRelayAddr2 = "/ip4/203.0.113.50/tcp/7777/p2p/12D3KooWQvzCBP1MdU6g3UC6rUwHtDkbMUWQKDapmHqQFPqZqTn7"

// --- v2 (default) tests ---

func TestV2EncodeDecodeRoundTrip(t *testing.T) {
	pid := generateTestPeerID(t)
	token, err := GenerateToken()
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}

	data := &InviteData{
		Token:     token,
		RelayAddr: testRelayAddr,
		PeerID:    pid,
	}

	code, err := Encode(data)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	t.Logf("v2 invite code (%d chars): %s", len(code), code)

	decoded, err := Decode(code)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}

	if decoded.Version != VersionV2 {
		t.Errorf("Version = %d, want %d", decoded.Version, VersionV2)
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
	if decoded.Network != "" {
		t.Errorf("Network should be empty for global, got %q", decoded.Network)
	}
}

func TestV2WithNamespace(t *testing.T) {
	pid := generateTestPeerID(t)
	token, _ := GenerateToken()

	data := &InviteData{
		Token:     token,
		RelayAddr: testRelayAddr,
		PeerID:    pid,
		Network:   "my-crew",
	}

	code, err := Encode(data)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	t.Logf("v2 invite code with namespace (%d chars): %s", len(code), code)

	decoded, err := Decode(code)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}

	if decoded.Version != VersionV2 {
		t.Errorf("Version = %d, want %d", decoded.Version, VersionV2)
	}
	if decoded.Network != "my-crew" {
		t.Errorf("Network = %q, want %q", decoded.Network, "my-crew")
	}
	if decoded.PeerID != pid {
		t.Errorf("PeerID mismatch")
	}
}

func TestV2NamespaceTooLong(t *testing.T) {
	pid := generateTestPeerID(t)
	token, _ := GenerateToken()

	data := &InviteData{
		Token:     token,
		RelayAddr: testRelayAddr,
		PeerID:    pid,
		Network:   strings.Repeat("a", 64), // exceeds 63-char limit
	}

	_, err := Encode(data)
	if err == nil {
		t.Fatal("Encode should reject namespace > 63 chars")
	}
}

// --- v1 backward compatibility tests ---

func TestV1EncodeDecodeRoundTrip(t *testing.T) {
	pid := generateTestPeerID(t)
	token, _ := GenerateToken()

	data := &InviteData{
		Version:   VersionV1,
		Token:     token,
		RelayAddr: testRelayAddr,
		PeerID:    pid,
	}

	code, err := Encode(data)
	if err != nil {
		t.Fatalf("Encode v1: %v", err)
	}

	decoded, err := Decode(code)
	if err != nil {
		t.Fatalf("Decode v1: %v", err)
	}

	if decoded.Version != VersionV1 {
		t.Errorf("Version = %d, want %d", decoded.Version, VersionV1)
	}
	if token != decoded.Token {
		t.Errorf("Token mismatch")
	}
	if data.RelayAddr != decoded.RelayAddr {
		t.Errorf("RelayAddr mismatch")
	}
	if data.PeerID != decoded.PeerID {
		t.Errorf("PeerID mismatch")
	}
	if decoded.Network != "" {
		t.Errorf("v1 should have empty network, got %q", decoded.Network)
	}
}

// --- Shared tests ---

func TestTokenHex(t *testing.T) {
	token, err := GenerateToken()
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}
	hex := TokenHex(token)
	if len(hex) != 16 { // 8 bytes = 16 hex chars
		t.Errorf("TokenHex length = %d, want 16", len(hex))
	}
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

func TestDecodeFutureVersion(t *testing.T) {
	// Version 0x03 should be rejected with a helpful message
	_, err := Decode("AYAA") // version byte 0x03
	if err == nil {
		t.Error("expected error for future version")
	}
	// The base32 might not decode to a valid version byte, but let's test
	// with a known prefix. Construct a raw payload with version 0x03.
	raw := make([]byte, 20)
	raw[0] = 0x03
	encoded := encoding.EncodeToString(raw)
	_, err = Decode(encoded)
	if err == nil {
		t.Error("expected error for version 3")
	}
	if !strings.Contains(err.Error(), "newer than supported") {
		t.Errorf("error should mention upgrade, got: %v", err)
	}
}

func TestDecodeRejectsTrailingJunk(t *testing.T) {
	pid := generateTestPeerID(t)
	token, err := GenerateToken()
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}

	data := &InviteData{
		Token:     token,
		RelayAddr: testRelayAddr2,
		PeerID:    pid,
	}

	code, err := Encode(data)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	// Simulate: fs.Args() = [code, "--name", "laptop"]
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

func TestV2GlobalNetworkOmitsNamespace(t *testing.T) {
	// When Network is empty, v2 encodes namespace length as 0.
	// Decoded code should have empty Network.
	pid := generateTestPeerID(t)
	token, _ := GenerateToken()

	data := &InviteData{
		Token:     token,
		RelayAddr: testRelayAddr,
		PeerID:    pid,
		Network:   "", // global
	}

	code, _ := Encode(data)
	decoded, err := Decode(code)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if decoded.Network != "" {
		t.Errorf("expected empty network for global, got %q", decoded.Network)
	}
}

func TestV2MaxNamespace(t *testing.T) {
	pid := generateTestPeerID(t)
	token, _ := GenerateToken()

	// 63 chars is the max (DNS label limit)
	ns := strings.Repeat("a", 63)
	data := &InviteData{
		Token:     token,
		RelayAddr: testRelayAddr,
		PeerID:    pid,
		Network:   ns,
	}

	code, err := Encode(data)
	if err != nil {
		t.Fatalf("Encode with 63-char namespace: %v", err)
	}

	decoded, err := Decode(code)
	if err != nil {
		t.Fatalf("Decode with 63-char namespace: %v", err)
	}
	if decoded.Network != ns {
		t.Errorf("namespace mismatch: got %q", decoded.Network)
	}
}
