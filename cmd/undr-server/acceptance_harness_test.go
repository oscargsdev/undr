package main

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const baseBinName = "temp-testbinary"

func LaunchTestProgram(port string) (cleanup func(), err error) {
	testDBDSN, err := acceptanceTestDSN()
	if err != nil {
		return nil, err
	}

	if err := setupAcceptanceTestDatabase(testDBDSN); err != nil {
		return nil, err
	}

	binPath, err := buildBinary()
	if err != nil {
		return nil, err
	}

	kill, err := runServer(binPath, port, testDBDSN)

	cleanup = func() {
		if kill != nil {
			kill()
		}

		if dbErr := cleanupAcceptanceTestDatabase(testDBDSN); dbErr != nil {
			fmt.Fprintf(os.Stderr, "acceptance test cleanup migration failed: %v\n", dbErr)
		}

		_ = os.Remove(binPath)
	}

	if err != nil {
		cleanup() // even though it's not listening correctly, the program could still be running
		return nil, err
	}

	return cleanup, nil
}

func buildBinary() (string, error) {
	// CreateTemp gives us a collision-free path; we remove the placeholder file
	// because `go build -o` writes the binary path itself.
	tmpFile, err := os.CreateTemp("", "*-"+baseBinName)
	if err != nil {
		return "", fmt.Errorf("cannot create temp file for test binary: %w", err)
	}

	binPath := tmpFile.Name()
	if err := tmpFile.Close(); err != nil {
		return "", fmt.Errorf("cannot close temp file for test binary: %w", err)
	}
	if err := os.Remove(binPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("cannot clear temp file before go build: %w", err)
	}

	build := exec.Command("go", "build", "-o", binPath)
	output, err := build.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("cannot build test binary %s: %w (output: %s)", binPath, err, strings.TrimSpace(string(output)))
	}

	return binPath, nil
}

func runServer(binPath string, port string, testDBDSN string) (kill func(), err error) {
	cmd := exec.Command(
		binPath,
		"-port", port,
		"-env", "test",
		"-db-dsn", testDBDSN,
		"-db-timeout", "1",
		"-identity-issuer", "undr-auth-test",
		"-jwt-expiration", "1",
		"-refresh-token-expiration", "1",
		"-activation-expiration", "1",
		"-db-max-open-conns", "5",
		"-db-max-idle-conns", "5",
	)

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("cannot run test server binary: %w", err)
	}

	kill = func() {
		_ = cmd.Process.Kill()
	}

	if err := waitForServerListening(port); err != nil {
		kill()
		return nil, err
	}

	return kill, nil
}

func waitForServerListening(port string) error {
	for range 30 {
		conn, err := net.DialTimeout("tcp", net.JoinHostPort("localhost", port), 100*time.Millisecond)
		if err == nil {
			_ = conn.Close()
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}

	return fmt.Errorf("nothing seems to be listening on localhost:%s", port)
}

func setupAcceptanceTestDatabase(testDBDSN string) error {
	// Setup uses down+up so tests always start from a clean schema/data state.
	if err := runMigrations(testDBDSN, "down"); err != nil {
		return fmt.Errorf("setup migrations down failed: %w", err)
	}

	if err := runMigrations(testDBDSN, "up"); err != nil {
		return fmt.Errorf("setup migrations up failed: %w", err)
	}

	return nil
}

func cleanupAcceptanceTestDatabase(testDBDSN string) error {
	// Cleanup runs down so acceptance tests do not leave persistent state behind.
	return runMigrations(testDBDSN, "down")
}

func runMigrations(testDBDSN string, direction string) error {
	if direction != "up" && direction != "down" {
		return fmt.Errorf("unknown migration direction %q", direction)
	}

	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	repoRoot, err := findRepoRoot(wd)
	if err != nil {
		return err
	}

	migrationsPath := filepath.Join(repoRoot, "migrations")

	args := []string{direction}
	if direction == "down" {
		args = append(args, "-all")
	}

	output, err := runMigrateCommand(migrationsPath, testDBDSN, args...)
	if err == nil {
		return nil
	}

	out := strings.TrimSpace(output)
	if strings.Contains(strings.ToLower(out), "no change") {
		return nil
	}

	// If a previous migration crashed, `migrate` marks DB as dirty.
	// For a dedicated acceptance-test DB we can safely drop and rebuild state.
	if strings.Contains(strings.ToLower(out), "dirty database version") {
		if _, dropErr := runMigrateCommand(migrationsPath, testDBDSN, "drop", "-f"); dropErr != nil {
			return fmt.Errorf("migrate %s failed with dirty database and drop reset failed: %w (output: %s)", direction, dropErr, out)
		}

		if direction == "up" {
			retryOutput, retryErr := runMigrateCommand(migrationsPath, testDBDSN, "up")
			if retryErr != nil && !strings.Contains(strings.ToLower(strings.TrimSpace(retryOutput)), "no change") {
				return fmt.Errorf("migrate up retry failed after dirty reset: %w (output: %s)", retryErr, strings.TrimSpace(retryOutput))
			}
		}

		return nil
	}

	return fmt.Errorf("migrate %s failed: %w (output: %s)", direction, err, out)
}

func runMigrateCommand(migrationsPath string, testDBDSN string, args ...string) (string, error) {
	baseArgs := []string{"-path", migrationsPath, "-database", testDBDSN}
	fullArgs := append(baseArgs, args...)

	cmd := exec.Command("migrate", fullArgs...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func acceptanceTestDSN() (string, error) {
	testDBDSN := strings.TrimSpace(os.Getenv("UNDR_TEST_DB_DSN"))
	if testDBDSN == "" {
		return "", errors.New("UNDR_TEST_DB_DSN env var is required for acceptance tests")
	}

	devDBDSN := strings.TrimSpace(os.Getenv("UNDR_DB_DSN"))
	if devDBDSN != "" && testDBDSN == devDBDSN {
		return "", errors.New("UNDR_TEST_DB_DSN must be different from UNDR_DB_DSN")
	}

	return testDBDSN, nil
}

func findRepoRoot(startDir string) (string, error) {
	dir := startDir

	for {
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("could not find repository root from %s", startDir)
		}
		dir = parent
	}
}
