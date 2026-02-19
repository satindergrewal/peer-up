package p2pnet

import (
	"errors"
	"testing"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
)

func genTestPeerID(t *testing.T) peer.ID {
	t.Helper()
	priv, _, err := crypto.GenerateKeyPair(crypto.Ed25519, -1)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	pid, err := peer.IDFromPrivateKey(priv)
	if err != nil {
		t.Fatalf("peer ID: %v", err)
	}
	return pid
}

func TestNewNameResolver(t *testing.T) {
	r := NewNameResolver()
	if r == nil {
		t.Fatal("NewNameResolver returned nil")
	}
	if len(r.List()) != 0 {
		t.Error("new resolver should have empty list")
	}
}

func TestNameResolverRegister(t *testing.T) {
	r := NewNameResolver()
	pid := genTestPeerID(t)

	// Valid registration
	if err := r.Register("home", pid); err != nil {
		t.Fatalf("Register: %v", err)
	}

	// Empty name should fail
	if err := r.Register("", pid); err == nil {
		t.Error("expected error for empty name")
	}

	// Overwrite with same name (allowed)
	pid2 := genTestPeerID(t)
	if err := r.Register("home", pid2); err != nil {
		t.Fatalf("Register overwrite: %v", err)
	}
	resolved, err := r.Resolve("home")
	if err != nil {
		t.Fatalf("Resolve after overwrite: %v", err)
	}
	if resolved != pid2 {
		t.Error("overwritten name should resolve to new peer ID")
	}
}

func TestNameResolverUnregister(t *testing.T) {
	r := NewNameResolver()
	pid := genTestPeerID(t)

	r.Register("home", pid)
	r.Unregister("home")

	_, err := r.Resolve("home")
	if err == nil {
		t.Error("expected error after unregister")
	}

	// Unregister non-existent name (no-op, no error)
	r.Unregister("nonexistent")
}

func TestNameResolverResolve(t *testing.T) {
	r := NewNameResolver()
	pid := genTestPeerID(t)
	r.Register("home", pid)

	t.Run("by name", func(t *testing.T) {
		resolved, err := r.Resolve("home")
		if err != nil {
			t.Fatalf("Resolve: %v", err)
		}
		if resolved != pid {
			t.Errorf("got %s, want %s", resolved, pid)
		}
	})

	t.Run("by peer ID string", func(t *testing.T) {
		resolved, err := r.Resolve(pid.String())
		if err != nil {
			t.Fatalf("Resolve by peer ID: %v", err)
		}
		if resolved != pid {
			t.Errorf("got %s, want %s", resolved, pid)
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, err := r.Resolve("nonexistent")
		if err == nil {
			t.Error("expected error for nonexistent name")
		}
		if !errors.Is(err, ErrNameNotFound) {
			t.Errorf("expected ErrNameNotFound, got: %v", err)
		}
	})
}

func TestNameResolverList(t *testing.T) {
	r := NewNameResolver()
	pid1 := genTestPeerID(t)
	pid2 := genTestPeerID(t)

	r.Register("home", pid1)
	r.Register("work", pid2)

	list := r.List()
	if len(list) != 2 {
		t.Fatalf("List() returned %d entries, want 2", len(list))
	}
	if list["home"] != pid1 {
		t.Errorf("home = %s, want %s", list["home"], pid1)
	}
	if list["work"] != pid2 {
		t.Errorf("work = %s, want %s", list["work"], pid2)
	}

	// Verify returned map is a copy (modifying it doesn't affect resolver)
	list["home"] = pid2
	list2 := r.List()
	if list2["home"] != pid1 {
		t.Error("List() should return a copy")
	}
}

func TestNameResolverLoadFromMap(t *testing.T) {
	r := NewNameResolver()
	pid := genTestPeerID(t)

	t.Run("valid", func(t *testing.T) {
		names := map[string]string{
			"home": pid.String(),
		}
		if err := r.LoadFromMap(names); err != nil {
			t.Fatalf("LoadFromMap: %v", err)
		}
		resolved, err := r.Resolve("home")
		if err != nil {
			t.Fatalf("Resolve: %v", err)
		}
		if resolved != pid {
			t.Errorf("got %s, want %s", resolved, pid)
		}
	})

	t.Run("invalid peer ID", func(t *testing.T) {
		names := map[string]string{
			"bad": "not-a-valid-peer-id",
		}
		if err := r.LoadFromMap(names); err == nil {
			t.Error("expected error for invalid peer ID")
		}
	})

	t.Run("empty map", func(t *testing.T) {
		if err := r.LoadFromMap(map[string]string{}); err != nil {
			t.Fatalf("LoadFromMap empty: %v", err)
		}
	})
}
