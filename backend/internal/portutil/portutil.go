package portutil

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// ProcessInfo holds information about a process occupying a port.
type ProcessInfo struct {
	PID     int
	Command string
}

// IsPortAvailable checks if a TCP port can be bound.
func IsPortAvailable(port string) bool {
	cleanPort := strings.TrimPrefix(port, ":")
	ln, err := net.Listen("tcp", ":"+cleanPort)
	if err != nil {
		return false
	}
	_ = ln.Close()
	return true
}

// FindProcessesOnPort returns PIDs of processes listening on the specified port.
// Excludes the current process PID.
func FindProcessesOnPort(port string) ([]ProcessInfo, error) {
	cleanPort := strings.TrimPrefix(port, ":")
	currentPID := os.Getpid()
	var results []ProcessInfo
	seenPIDs := make(map[int]bool)

	if runtime.GOOS == "windows" {
		// Windows: use netstat -ano
		cmd := exec.Command("cmd.exe", "/C", fmt.Sprintf("netstat -ano | findstr :%s", cleanPort))
		out, err := cmd.Output()
		if err == nil {
			lines := strings.Split(string(out), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if !strings.Contains(line, "LISTENING") {
					continue
				}
				fields := strings.Fields(line)
				if len(fields) >= 5 {
					pidStr := fields[len(fields)-1]
					pid, convErr := strconv.Atoi(pidStr)
					if convErr == nil && pid > 0 && pid != currentPID && !seenPIDs[pid] {
						seenPIDs[pid] = true
						results = append(results, ProcessInfo{PID: pid, Command: "windows-process"})
					}
				}
			}
		}
		return results, nil
	}

	// Linux / Unix: Primary strategy via lsof
	lsofPath, err := exec.LookPath("lsof")
	if err == nil {
		cmd := exec.Command(lsofPath, "-t", "-i", fmt.Sprintf(":%s", cleanPort))
		out, err := cmd.Output()
		if err == nil {
			for _, line := range strings.Split(string(out), "\n") {
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}
				pid, convErr := strconv.Atoi(line)
				if convErr == nil && pid > 0 && pid != currentPID && !seenPIDs[pid] {
					seenPIDs[pid] = true
					cmdName := getProcessCommand(pid)
					results = append(results, ProcessInfo{PID: pid, Command: cmdName})
				}
			}
		}
	}

	// Secondary strategy via fuser if lsof returned nothing
	if len(results) == 0 {
		fuserPath, err := exec.LookPath("fuser")
		if err == nil {
			cmd := exec.Command(fuserPath, fmt.Sprintf("%s/tcp", cleanPort))
			var outBuf, errBuf bytes.Buffer
			cmd.Stdout = &outBuf
			cmd.Stderr = &errBuf
			_ = cmd.Run()

			// fuser outputs PIDs to stdout or stderr depending on distro
			combined := outBuf.String() + " " + errBuf.String()
			fields := strings.Fields(combined)
			for _, field := range fields {
				field = strings.TrimSuffix(strings.TrimSpace(field), "/tcp:")
				field = strings.TrimPrefix(field, ":")
				pid, convErr := strconv.Atoi(field)
				if convErr == nil && pid > 0 && pid != currentPID && !seenPIDs[pid] {
					seenPIDs[pid] = true
					cmdName := getProcessCommand(pid)
					results = append(results, ProcessInfo{PID: pid, Command: cmdName})
				}
			}
		}
	}

	// Tertiary strategy via ss if still nothing
	if len(results) == 0 {
		ssPath, err := exec.LookPath("ss")
		if err == nil {
			cmd := exec.Command(ssPath, "-tulpn", fmt.Sprintf("( sport = :%s )", cleanPort))
			out, err := cmd.Output()
			if err == nil {
				// Search for pid=XXXX
				str := string(out)
				for {
					idx := strings.Index(str, "pid=")
					if idx == -1 {
						break
					}
					str = str[idx+4:]
					endIdx := strings.IndexAny(str, ",)")
					if endIdx != -1 {
						pidStr := str[:endIdx]
						pid, convErr := strconv.Atoi(pidStr)
						if convErr == nil && pid > 0 && pid != currentPID && !seenPIDs[pid] {
							seenPIDs[pid] = true
							cmdName := getProcessCommand(pid)
							results = append(results, ProcessInfo{PID: pid, Command: cmdName})
						}
					}
				}
			}
		}
	}

	return results, nil
}

// getProcessCommand attempts to read the process name from /proc/<pid>/comm or /proc/<pid>/cmdline.
func getProcessCommand(pid int) string {
	commPath := fmt.Sprintf("/proc/%d/comm", pid)
	data, err := os.ReadFile(commPath)
	if err == nil {
		return strings.TrimSpace(string(data))
	}
	cmdlinePath := fmt.Sprintf("/proc/%d/cmdline", pid)
	data, err = os.ReadFile(cmdlinePath)
	if err == nil {
		parts := strings.Split(string(data), "\x00")
		if len(parts) > 0 && parts[0] != "" {
			return parts[0]
		}
	}
	return fmt.Sprintf("PID:%d", pid)
}

// TerminateProcess attempts graceful termination, followed by force kill if needed.
func TerminateProcess(pid int) error {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}

	if runtime.GOOS == "windows" {
		cmd := exec.Command("taskkill", "/F", "/PID", strconv.Itoa(pid))
		return cmd.Run()
	}

	// On POSIX: Send SIGTERM first for graceful cleanup
	_ = proc.Signal(syscall.SIGTERM)

	// Wait up to 300ms to allow graceful exit
	deadline := time.Now().Add(300 * time.Millisecond)
	for time.Now().Before(deadline) {
		// Check if process still exists
		if err := proc.Signal(syscall.Signal(0)); err != nil {
			return nil // Exited
		}
		time.Sleep(50 * time.Millisecond)
	}

	// If still alive, send SIGKILL
	if err := proc.Signal(syscall.SIGKILL); err != nil {
		// If error is process already dead, that's fine
		if strings.Contains(err.Error(), "already finished") || strings.Contains(err.Error(), "no such process") {
			return nil
		}
		return err
	}
	return nil
}

// ClearPort checks if the port is in use, identifies occupying processes, terminates them,
// and waits until the port is free or timeout expires.
func ClearPort(port string, timeout time.Duration) error {
	cleanPort := strings.TrimPrefix(port, ":")

	if IsPortAvailable(cleanPort) {
		return nil
	}

	procs, err := FindProcessesOnPort(cleanPort)
	if err != nil {
		return fmt.Errorf("failed to discover processes on port %s: %w", cleanPort, err)
	}

	if len(procs) == 0 {
		// Port is not bindable, but no PIDs discovered via tools.
		// Try a direct fuser kill fallback if available on Linux.
		if runtime.GOOS != "windows" {
			fuserPath, err := exec.LookPath("fuser")
			if err == nil {
				_ = exec.Command(fuserPath, "-k", "-9", fmt.Sprintf("%s/tcp", cleanPort)).Run()
			}
		}
	} else {
		for _, p := range procs {
			log.Printf("⚠️ Found contentious process [PID: %d | %s] on port %s. Clearing...", p.PID, p.Command, cleanPort)
			if termErr := TerminateProcess(p.PID); termErr != nil {
				log.Printf("Warning: failed to terminate contentious PID %d: %v", p.PID, termErr)
			}
		}
	}

	// Poll until port becomes available or timeout expires
	start := time.Now()
	for time.Since(start) < timeout {
		if IsPortAvailable(cleanPort) {
			return nil
		}
		time.Sleep(50 * time.Millisecond)
	}

	return fmt.Errorf("port %s remained contentious after %v timeout", cleanPort, timeout)
}

// ClearPortIfContentious inspects the given port at startup. If occupied, it automatically
// remediates port contention and reclaims the socket.
func ClearPortIfContentious(port string) error {
	cleanPort := strings.TrimPrefix(port, ":")

	// Allow opt-out via environment variable if desired
	if val := os.Getenv("MERCURY_CLEAR_PORT"); strings.EqualFold(val, "false") || val == "0" {
		log.Printf("ℹ Port clearing at startup skipped (MERCURY_CLEAR_PORT=%s)", val)
		return nil
	}

	if IsPortAvailable(cleanPort) {
		log.Printf("✔ Port :%s is free and available.", cleanPort)
		return nil
	}

	log.Printf("⚠️ Port :%s is currently occupied! Resolving port contention at startup...", cleanPort)
	err := ClearPort(cleanPort, 3*time.Second)
	if err != nil {
		return fmt.Errorf("could not clear contentious port :%s: %w", cleanPort, err)
	}

	log.Printf("✔ Contentious port :%s successfully cleared and reclaimed.", cleanPort)
	return nil
}
