package main

import (
	"io"
	"os"
	"strings"
	"testing"
)

// captureExit overrides the package-level osExit variable so that calls to
// osExit inside fn are intercepted.  It returns the exit code and a boolean
// indicating whether osExit was actually called.
//
// How it works: the replacement panics with an exitSentinel value — the same
// type defined in exit.go — which immediately unwinds the call stack (just
// like a real os.Exit would halt the process).  A deferred recover catches
// the sentinel and stores the code.  Any other panic is re-raised.
func captureExit(fn func()) (code int, exited bool) {
	old := osExit
	defer func() { osExit = old }()

	osExit = func(c int) {
		panic(exitSentinel(c))
	}

	func() {
		defer func() {
			if r := recover(); r != nil {
				if s, ok := r.(exitSentinel); ok {
					code = int(s)
					exited = true
				} else {
					panic(r) // re-raise non-sentinel panics
				}
			}
		}()
		fn()
	}()
	return code, exited
}

// captureStderr redirects os.Stderr during fn and returns what was written.
func captureStderr(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stderr = w

	fn()

	w.Close()
	os.Stderr = old
	data, _ := io.ReadAll(r)
	return string(data)
}

// ---------------------------------------------------------------------------
// Category 1: Thin runXxx wrappers that call doXxx → osExit(1) on error.
//
// We test the ERROR path by passing --config pointing at a nonexistent file.
// Each wrapper's doXxx function will fail with "config error", the wrapper
// prints "Error: ..." to stderr, and calls osExit(1).
//
// For the SUCCESS path, the doXxx functions already have thorough tests.
// Here we verify the wrapper's plumbing: doXxx success → no osExit.
// ---------------------------------------------------------------------------

func TestRunConfigValidate_Error(t *testing.T) {
	code, exited := captureExit(func() {
		runConfigValidate([]string{"--config", "/tmp/nonexistent-peerup-test/peerup.yaml"})
	})
	if !exited || code != 1 {
		t.Errorf("expected exit(1), got exited=%v code=%d", exited, code)
	}
}

func TestRunConfigShow_Error(t *testing.T) {
	code, exited := captureExit(func() {
		runConfigShow([]string{"--config", "/tmp/nonexistent-peerup-test/peerup.yaml"})
	})
	if !exited || code != 1 {
		t.Errorf("expected exit(1), got exited=%v code=%d", exited, code)
	}
}

func TestRunConfigRollback_Error(t *testing.T) {
	code, exited := captureExit(func() {
		runConfigRollback([]string{"--config", "/tmp/nonexistent-peerup-test/peerup.yaml"})
	})
	if !exited || code != 1 {
		t.Errorf("expected exit(1), got exited=%v code=%d", exited, code)
	}
}

func TestRunConfigApply_Error(t *testing.T) {
	code, exited := captureExit(func() {
		runConfigApply([]string{"--config", "/tmp/nonexistent-peerup-test/peerup.yaml"})
	})
	if !exited || code != 1 {
		t.Errorf("expected exit(1), got exited=%v code=%d", exited, code)
	}
}

func TestRunConfigConfirm_Error(t *testing.T) {
	code, exited := captureExit(func() {
		runConfigConfirm([]string{"--config", "/tmp/nonexistent-peerup-test/peerup.yaml"})
	})
	if !exited || code != 1 {
		t.Errorf("expected exit(1), got exited=%v code=%d", exited, code)
	}
}

func TestRunAuthAdd_Error(t *testing.T) {
	code, exited := captureExit(func() {
		runAuthAdd([]string{"--config", "/tmp/nonexistent-peerup-test/peerup.yaml", "fake-id"})
	})
	if !exited || code != 1 {
		t.Errorf("expected exit(1), got exited=%v code=%d", exited, code)
	}
}

func TestRunAuthList_Error(t *testing.T) {
	code, exited := captureExit(func() {
		runAuthList([]string{"--config", "/tmp/nonexistent-peerup-test/peerup.yaml"})
	})
	if !exited || code != 1 {
		t.Errorf("expected exit(1), got exited=%v code=%d", exited, code)
	}
}

func TestRunAuthRemove_Error(t *testing.T) {
	code, exited := captureExit(func() {
		runAuthRemove([]string{"--config", "/tmp/nonexistent-peerup-test/peerup.yaml", "fake-id"})
	})
	if !exited || code != 1 {
		t.Errorf("expected exit(1), got exited=%v code=%d", exited, code)
	}
}

func TestRunAuthValidate_Error(t *testing.T) {
	code, exited := captureExit(func() {
		runAuthValidate([]string{"--config", "/tmp/nonexistent-peerup-test/peerup.yaml"})
	})
	if !exited || code != 1 {
		t.Errorf("expected exit(1), got exited=%v code=%d", exited, code)
	}
}

func TestRunRelayAdd_Error(t *testing.T) {
	code, exited := captureExit(func() {
		runRelayAdd([]string{"--config", "/tmp/nonexistent-peerup-test/peerup.yaml", "/ip4/1.2.3.4/tcp/7777"})
	})
	if !exited || code != 1 {
		t.Errorf("expected exit(1), got exited=%v code=%d", exited, code)
	}
}

func TestRunRelayList_Error(t *testing.T) {
	code, exited := captureExit(func() {
		runRelayList([]string{"--config", "/tmp/nonexistent-peerup-test/peerup.yaml"})
	})
	if !exited || code != 1 {
		t.Errorf("expected exit(1), got exited=%v code=%d", exited, code)
	}
}

func TestRunRelayRemove_Error(t *testing.T) {
	code, exited := captureExit(func() {
		runRelayRemove([]string{"--config", "/tmp/nonexistent-peerup-test/peerup.yaml", "/ip4/1.2.3.4/tcp/7777"})
	})
	if !exited || code != 1 {
		t.Errorf("expected exit(1), got exited=%v code=%d", exited, code)
	}
}

func TestRunServiceAdd_Error(t *testing.T) {
	code, exited := captureExit(func() {
		runServiceAdd([]string{"--config", "/tmp/nonexistent-peerup-test/peerup.yaml", "ssh", "localhost:22"})
	})
	if !exited || code != 1 {
		t.Errorf("expected exit(1), got exited=%v code=%d", exited, code)
	}
}

func TestRunServiceList_Error(t *testing.T) {
	code, exited := captureExit(func() {
		runServiceList([]string{"--config", "/tmp/nonexistent-peerup-test/peerup.yaml"})
	})
	if !exited || code != 1 {
		t.Errorf("expected exit(1), got exited=%v code=%d", exited, code)
	}
}

func TestRunServiceRemove_Error(t *testing.T) {
	code, exited := captureExit(func() {
		runServiceRemove([]string{"--config", "/tmp/nonexistent-peerup-test/peerup.yaml", "ssh"})
	})
	if !exited || code != 1 {
		t.Errorf("expected exit(1), got exited=%v code=%d", exited, code)
	}
}

func TestRunServiceEnable_Error(t *testing.T) {
	code, exited := captureExit(func() {
		runServiceSetEnabled([]string{"--config", "/tmp/nonexistent-peerup-test/peerup.yaml", "ssh"}, true)
	})
	if !exited || code != 1 {
		t.Errorf("expected exit(1), got exited=%v code=%d", exited, code)
	}
}

func TestRunServiceDisable_Error(t *testing.T) {
	code, exited := captureExit(func() {
		runServiceSetEnabled([]string{"--config", "/tmp/nonexistent-peerup-test/peerup.yaml", "ssh"}, false)
	})
	if !exited || code != 1 {
		t.Errorf("expected exit(1), got exited=%v code=%d", exited, code)
	}
}

func TestRunResolve_Error(t *testing.T) {
	code, exited := captureExit(func() {
		runResolve([]string{"--config", "/tmp/nonexistent-peerup-test/peerup.yaml", "some-peer"})
	})
	if !exited || code != 1 {
		t.Errorf("expected exit(1), got exited=%v code=%d", exited, code)
	}
}

func TestRunWhoami_Error(t *testing.T) {
	code, exited := captureExit(func() {
		runWhoami([]string{"--config", "/tmp/nonexistent-peerup-test/peerup.yaml"})
	})
	if !exited || code != 1 {
		t.Errorf("expected exit(1), got exited=%v code=%d", exited, code)
	}
}

func TestRunStatus_Error(t *testing.T) {
	code, exited := captureExit(func() {
		runStatus([]string{"--config", "/tmp/nonexistent-peerup-test/peerup.yaml"})
	})
	if !exited || code != 1 {
		t.Errorf("expected exit(1), got exited=%v code=%d", exited, code)
	}
}

func TestRunRelayAuthorize_Error(t *testing.T) {
	code, exited := captureExit(func() {
		runRelayAuthorize([]string{"fake-id"}, "/tmp/nonexistent-peerup-test/relay-server.yaml")
	})
	if !exited || code != 1 {
		t.Errorf("expected exit(1), got exited=%v code=%d", exited, code)
	}
}

func TestRunRelayDeauthorize_Error(t *testing.T) {
	code, exited := captureExit(func() {
		runRelayDeauthorize([]string{"fake-id"}, "/tmp/nonexistent-peerup-test/relay-server.yaml")
	})
	if !exited || code != 1 {
		t.Errorf("expected exit(1), got exited=%v code=%d", exited, code)
	}
}

func TestRunRelayListPeers_Error(t *testing.T) {
	code, exited := captureExit(func() {
		runRelayListPeers("/tmp/nonexistent-peerup-test/relay-server.yaml")
	})
	if !exited || code != 1 {
		t.Errorf("expected exit(1), got exited=%v code=%d", exited, code)
	}
}

func TestRunRelayServerConfigValidate_Error(t *testing.T) {
	code, exited := captureExit(func() {
		runRelayServerConfigValidate("/tmp/nonexistent-peerup-test/relay-server.yaml")
	})
	if !exited || code != 1 {
		t.Errorf("expected exit(1), got exited=%v code=%d", exited, code)
	}
}

func TestRunRelayServerConfigRollback_Error(t *testing.T) {
	code, exited := captureExit(func() {
		runRelayServerConfigRollback("/tmp/nonexistent-peerup-test/relay-server.yaml")
	})
	if !exited || code != 1 {
		t.Errorf("expected exit(1), got exited=%v code=%d", exited, code)
	}
}

// ---------------------------------------------------------------------------
// Category 1 SUCCESS paths: thin wrappers that should NOT call osExit.
// We only test a few representative ones — the doXxx functions themselves
// are exhaustively tested in their own test files.
// ---------------------------------------------------------------------------

func TestRunConfigValidate_Success(t *testing.T) {
	cfgPath := writeTestConfigDir(t)

	code, exited := captureExit(func() {
		runConfigValidate([]string{"--config", cfgPath})
	})
	if exited {
		t.Errorf("should not have exited, got code=%d", code)
	}
}

func TestRunConfigShow_Success(t *testing.T) {
	cfgPath := writeTestConfigDir(t)

	code, exited := captureExit(func() {
		runConfigShow([]string{"--config", cfgPath})
	})
	if exited {
		t.Errorf("should not have exited, got code=%d", code)
	}
}

func TestRunRelayList_Success(t *testing.T) {
	cfgPath := writeTestConfigDir(t)

	code, exited := captureExit(func() {
		runRelayList([]string{"--config", cfgPath})
	})
	if exited {
		t.Errorf("should not have exited, got code=%d", code)
	}
}

func TestRunServiceList_Success(t *testing.T) {
	cfgPath := writeTestConfigDir(t)

	code, exited := captureExit(func() {
		runServiceList([]string{"--config", cfgPath})
	})
	if exited {
		t.Errorf("should not have exited, got code=%d", code)
	}
}

func TestRunWhoami_Success(t *testing.T) {
	cfgPath := writeTestConfigDir(t)

	code, exited := captureExit(func() {
		runWhoami([]string{"--config", cfgPath})
	})
	if exited {
		t.Errorf("should not have exited, got code=%d", code)
	}
}

func TestRunStatus_Success(t *testing.T) {
	cfgPath := writeTestConfigDir(t)

	code, exited := captureExit(func() {
		runStatus([]string{"--config", cfgPath})
	})
	if exited {
		t.Errorf("should not have exited, got code=%d", code)
	}
}

func TestRunAuthList_Success(t *testing.T) {
	cfgPath := writeTestConfigDir(t)

	code, exited := captureExit(func() {
		runAuthList([]string{"--config", cfgPath})
	})
	if exited {
		t.Errorf("should not have exited, got code=%d", code)
	}
}

func TestRunConfigConfirm_Success_NoPending(t *testing.T) {
	cfgPath := writeTestConfigDir(t)

	code, exited := captureExit(func() {
		runConfigConfirm([]string{"--config", cfgPath})
	})
	// No pending commit-confirmed → doConfigConfirm returns error → exit(1)
	if !exited || code != 1 {
		t.Errorf("expected exit(1) for no pending config, got exited=%v code=%d", exited, code)
	}
}

// ---------------------------------------------------------------------------
// Category 2: Dispatchers — test unknown subcommand → osExit(1) and
// empty args → osExit(1).
// ---------------------------------------------------------------------------

func TestRunConfig_EmptyArgs(t *testing.T) {
	code, exited := captureExit(func() {
		runConfig(nil)
	})
	if !exited || code != 1 {
		t.Errorf("expected exit(1), got exited=%v code=%d", exited, code)
	}
}

func TestRunConfig_UnknownSubcommand(t *testing.T) {
	stderr := captureStderr(t, func() {
		code, exited := captureExit(func() {
			runConfig([]string{"bogus"})
		})
		if !exited || code != 1 {
			t.Errorf("expected exit(1), got exited=%v code=%d", exited, code)
		}
	})
	if !strings.Contains(stderr, "Unknown config command") {
		t.Errorf("stderr should mention unknown command, got: %s", stderr)
	}
}

func TestRunAuth_EmptyArgs(t *testing.T) {
	code, exited := captureExit(func() {
		runAuth(nil)
	})
	if !exited || code != 1 {
		t.Errorf("expected exit(1), got exited=%v code=%d", exited, code)
	}
}

func TestRunAuth_UnknownSubcommand(t *testing.T) {
	stderr := captureStderr(t, func() {
		code, exited := captureExit(func() {
			runAuth([]string{"bogus"})
		})
		if !exited || code != 1 {
			t.Errorf("expected exit(1), got exited=%v code=%d", exited, code)
		}
	})
	if !strings.Contains(stderr, "Unknown auth command") {
		t.Errorf("stderr should mention unknown command, got: %s", stderr)
	}
}

func TestRunService_EmptyArgs(t *testing.T) {
	code, exited := captureExit(func() {
		runService(nil)
	})
	if !exited || code != 1 {
		t.Errorf("expected exit(1), got exited=%v code=%d", exited, code)
	}
}

func TestRunService_UnknownSubcommand(t *testing.T) {
	stderr := captureStderr(t, func() {
		code, exited := captureExit(func() {
			runService([]string{"bogus"})
		})
		if !exited || code != 1 {
			t.Errorf("expected exit(1), got exited=%v code=%d", exited, code)
		}
	})
	if !strings.Contains(stderr, "Unknown service command") {
		t.Errorf("stderr should mention unknown command, got: %s", stderr)
	}
}

func TestRunRelay_EmptyArgs(t *testing.T) {
	code, exited := captureExit(func() {
		runRelay(nil)
	})
	if !exited || code != 1 {
		t.Errorf("expected exit(1), got exited=%v code=%d", exited, code)
	}
}

func TestRunRelay_UnknownSubcommand(t *testing.T) {
	stderr := captureStderr(t, func() {
		code, exited := captureExit(func() {
			runRelay([]string{"bogus"})
		})
		if !exited || code != 1 {
			t.Errorf("expected exit(1), got exited=%v code=%d", exited, code)
		}
	})
	if !strings.Contains(stderr, "Unknown relay command") {
		t.Errorf("stderr should mention unknown command, got: %s", stderr)
	}
}

func TestRunDaemon_UnknownSubcommand(t *testing.T) {
	stderr := captureStderr(t, func() {
		code, exited := captureExit(func() {
			runDaemon([]string{"bogus"})
		})
		if !exited || code != 1 {
			t.Errorf("expected exit(1), got exited=%v code=%d", exited, code)
		}
	})
	if !strings.Contains(stderr, "Unknown daemon subcommand") {
		t.Errorf("stderr should mention unknown subcommand, got: %s", stderr)
	}
}

func TestRunRelayServerConfig_EmptyArgs(t *testing.T) {
	code, exited := captureExit(func() {
		runRelayServerConfig(nil, "/tmp/nonexistent-peerup-test/relay-server.yaml")
	})
	if !exited || code != 1 {
		t.Errorf("expected exit(1), got exited=%v code=%d", exited, code)
	}
}

func TestRunRelayServerConfig_UnknownSubcommand(t *testing.T) {
	stderr := captureStderr(t, func() {
		code, exited := captureExit(func() {
			runRelayServerConfig([]string{"bogus"}, "/tmp/nonexistent-peerup-test/relay-server.yaml")
		})
		if !exited || code != 1 {
			t.Errorf("expected exit(1), got exited=%v code=%d", exited, code)
		}
	})
	if !strings.Contains(stderr, "Unknown config command") {
		t.Errorf("stderr should mention unknown command, got: %s", stderr)
	}
}

// ---------------------------------------------------------------------------
// Category 3: printXxxUsage functions — just verify they don't panic.
// ---------------------------------------------------------------------------

func TestPrintUsage(t *testing.T) {
	// Redirect stdout to discard (these write to os.Stdout directly)
	old := os.Stdout
	os.Stdout = os.NewFile(0, os.DevNull)
	defer func() { os.Stdout = old }()

	printUsage()
}

func TestPrintVersion(t *testing.T) {
	old := os.Stdout
	os.Stdout = os.NewFile(0, os.DevNull)
	defer func() { os.Stdout = old }()

	printVersion()
}

func TestPrintDaemonUsage(t *testing.T) {
	old := os.Stdout
	os.Stdout = os.NewFile(0, os.DevNull)
	defer func() { os.Stdout = old }()

	printDaemonUsage()
}

func TestPrintAuthUsage(t *testing.T) {
	old := os.Stdout
	os.Stdout = os.NewFile(0, os.DevNull)
	defer func() { os.Stdout = old }()

	printAuthUsage()
}

func TestPrintConfigUsage(t *testing.T) {
	old := os.Stdout
	os.Stdout = os.NewFile(0, os.DevNull)
	defer func() { os.Stdout = old }()

	printConfigUsage()
}

func TestPrintServiceUsage(t *testing.T) {
	old := os.Stdout
	os.Stdout = os.NewFile(0, os.DevNull)
	defer func() { os.Stdout = old }()

	printServiceUsage()
}

func TestPrintRelayServeUsage(t *testing.T) {
	old := os.Stdout
	os.Stdout = os.NewFile(0, os.DevNull)
	defer func() { os.Stdout = old }()

	printRelayServeUsage()
}

// ---------------------------------------------------------------------------
// Category 4: runRelayServerVersion — pure output, no exit.
// ---------------------------------------------------------------------------

func TestRunRelayServerVersion(t *testing.T) {
	old := os.Stdout
	os.Stdout = os.NewFile(0, os.DevNull)
	defer func() { os.Stdout = old }()

	code, exited := captureExit(func() {
		runRelayServerVersion()
	})
	if exited {
		t.Errorf("should not have exited, got code=%d", code)
	}
}

// ---------------------------------------------------------------------------
// Category 5: Daemon client commands — these call daemonClient() which will
// fail because no daemon is running.  Verify they osExit(1).
// ---------------------------------------------------------------------------

func TestRunDaemonStop_NoDaemon(t *testing.T) {
	code, exited := captureExit(func() {
		runDaemonStop()
	})
	if !exited || code != 1 {
		t.Errorf("expected exit(1) when no daemon, got exited=%v code=%d", exited, code)
	}
}

func TestRunDaemonStatus_NoDaemon(t *testing.T) {
	code, exited := captureExit(func() {
		runDaemonStatus(nil)
	})
	if !exited || code != 1 {
		t.Errorf("expected exit(1) when no daemon, got exited=%v code=%d", exited, code)
	}
}

func TestRunDaemonPing_NoArgs(t *testing.T) {
	code, exited := captureExit(func() {
		runDaemonPing(nil)
	})
	if !exited || code != 1 {
		t.Errorf("expected exit(1) for missing peer arg, got exited=%v code=%d", exited, code)
	}
}

func TestRunDaemonConnect_MissingFlags(t *testing.T) {
	code, exited := captureExit(func() {
		runDaemonConnect(nil)
	})
	if !exited || code != 1 {
		t.Errorf("expected exit(1) for missing flags, got exited=%v code=%d", exited, code)
	}
}

func TestRunDaemonDisconnect_NoArgs(t *testing.T) {
	code, exited := captureExit(func() {
		runDaemonDisconnect(nil)
	})
	if !exited || code != 1 {
		t.Errorf("expected exit(1) for missing proxy ID, got exited=%v code=%d", exited, code)
	}
}

func TestDaemonClient_NoDaemon(t *testing.T) {
	code, exited := captureExit(func() {
		_ = daemonClient()
	})
	if !exited || code != 1 {
		t.Errorf("expected exit(1) when no daemon, got exited=%v code=%d", exited, code)
	}
}

// ---------------------------------------------------------------------------
// Category 6: Relay dispatcher routes known subcommands correctly.
// Verify 'relay version' doesn't exit (it's pure output).
// ---------------------------------------------------------------------------

func TestRunRelay_Version(t *testing.T) {
	old := os.Stdout
	os.Stdout = os.NewFile(0, os.DevNull)
	defer func() { os.Stdout = old }()

	code, exited := captureExit(func() {
		runRelay([]string{"version"})
	})
	if exited {
		t.Errorf("relay version should not exit, got code=%d", code)
	}
}

// Verify dispatcher routes to runRelayAdd (which will error on bad config).
func TestRunRelay_Add_Dispatches(t *testing.T) {
	code, exited := captureExit(func() {
		runRelay([]string{"add", "--config", "/tmp/nonexistent-peerup-test/peerup.yaml", "/ip4/1.2.3.4/tcp/7777"})
	})
	if !exited || code != 1 {
		t.Errorf("expected exit(1), got exited=%v code=%d", exited, code)
	}
}

// Verify dispatcher routes to runRelayList (which will error on bad config).
func TestRunRelay_List_Dispatches(t *testing.T) {
	code, exited := captureExit(func() {
		runRelay([]string{"list", "--config", "/tmp/nonexistent-peerup-test/peerup.yaml"})
	})
	if !exited || code != 1 {
		t.Errorf("expected exit(1), got exited=%v code=%d", exited, code)
	}
}

// Verify dispatcher routes to runRelayRemove (which will error on bad config).
func TestRunRelay_Remove_Dispatches(t *testing.T) {
	code, exited := captureExit(func() {
		runRelay([]string{"remove", "--config", "/tmp/nonexistent-peerup-test/peerup.yaml", "/ip4/1.2.3.4/tcp/7777"})
	})
	if !exited || code != 1 {
		t.Errorf("expected exit(1), got exited=%v code=%d", exited, code)
	}
}
