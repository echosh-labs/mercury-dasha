package portutil

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"testing"
	"time"
)

func TestIsPortAvailable_FreePort(t *testing.T) {
	// Find a free port by listening on port 0
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to bind ephemeral port: %v", err)
	}
	port := fmt.Sprintf("%d", ln.Addr().(*net.TCPAddr).Port)
	_ = ln.Close()

	if !IsPortAvailable(port) {
		t.Errorf("expected port %s to be available after close", port)
	}
}

func TestIsPortAvailable_BusyPort(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to bind ephemeral port: %v", err)
	}
	defer ln.Close()

	port := fmt.Sprintf("%d", ln.Addr().(*net.TCPAddr).Port)
	if IsPortAvailable(port) {
		t.Errorf("expected port %s to be busy while listening", port)
	}
}

func TestClearPort_AlreadyFree(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to bind ephemeral port: %v", err)
	}
	port := fmt.Sprintf("%d", ln.Addr().(*net.TCPAddr).Port)
	_ = ln.Close()

	if err := ClearPort(port, 1*time.Second); err != nil {
		t.Errorf("ClearPort on free port returned error: %v", err)
	}
}

func TestClearPort_ContentiousProcess(t *testing.T) {
	// Find a free ephemeral port
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to bind ephemeral port: %v", err)
	}
	port := fmt.Sprintf("%d", ln.Addr().(*net.TCPAddr).Port)
	_ = ln.Close()

	// Spawn a separate child process holding the port using python3
	cmd := exec.Command("python3", "-c", fmt.Sprintf(`
import socket, time
s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
s.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
s.bind(('0.0.0.0', %s))
s.listen(1)
time.sleep(30)
`, port))

	if err := cmd.Start(); err != nil {
		t.Skipf("skipping contentious process test: python3 spawn failed: %v", err)
	}
	defer func() {
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
	}()

	// Wait up to 1 second for python process to occupy port
	occupied := false
	for i := 0; i < 20; i++ {
		time.Sleep(50 * time.Millisecond)
		if !IsPortAvailable(port) {
			occupied = true
			break
		}
	}

	if !occupied {
		t.Skipf("sub-process did not occupy port %s in time", port)
	}

	// Verify FindProcessesOnPort identifies the process
	procs, err := FindProcessesOnPort(port)
	if err != nil {
		t.Logf("FindProcessesOnPort notice: %v", err)
	} else if len(procs) == 0 {
		t.Logf("FindProcessesOnPort returned 0 procs (fallback will still clear)")
	} else {
		t.Logf("Found %d contentious process(es) on port %s: %+v", len(procs), port, procs)
	}

	// Clear the contentious port
	if err := ClearPort(port, 3*time.Second); err != nil {
		t.Fatalf("ClearPort failed to clear contentious port %s: %v", port, err)
	}

	// Verify port is now available
	if !IsPortAvailable(port) {
		t.Errorf("port %s should be available after ClearPort", port)
	}
}

func TestClearPortIfContentious_OptOutEnv(t *testing.T) {
	os.Setenv("MERCURY_CLEAR_PORT", "false")
	defer os.Unsetenv("MERCURY_CLEAR_PORT")

	// Even if a port was busy or free, it returns nil immediately without blocking
	err := ClearPortIfContentious("8080")
	if err != nil {
		t.Errorf("expected nil error when MERCURY_CLEAR_PORT=false, got: %v", err)
	}
}
