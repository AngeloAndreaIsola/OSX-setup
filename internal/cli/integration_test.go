//go:build integration

package cli_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestE2EIntegration(t *testing.T) {
	// Step 1: Create a temporary directory for isolation
	tempDir, err := os.MkdirTemp("", "setupper-integration-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Build the setupper binary
	binaryPath := filepath.Join(tempDir, "setupper")
	buildCmd := exec.Command("go", "build", "-o", binaryPath, "../../cmd/setupper")
	var buildStderr bytes.Buffer
	buildCmd.Stderr = &buildStderr
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("failed to build setupper: %v\nStderr: %s", err, buildStderr.String())
	}

	// Step 2: Set up environment variables to isolate from the developer's real system
	gitConfigPath := filepath.Join(tempDir, ".gitconfig")
	zshrcPath := filepath.Join(tempDir, ".zshrc")
	
	// Pre-create a blank .zshrc
	if err := os.WriteFile(zshrcPath, []byte("# Empty zshrc for testing\n"), 0644); err != nil {
		t.Fatalf("failed to pre-create .zshrc: %v", err)
	}

	// Pre-seed some mock git configurations in our isolated git config file
	gitSeedCmd := exec.Command("git", "config", "--file", gitConfigPath, "user.name", "E2E Integration Test User")
	if err := gitSeedCmd.Run(); err != nil {
		t.Fatalf("failed to seed user.name in temp git config: %v", err)
	}
	gitSeedCmd2 := exec.Command("git", "config", "--file", gitConfigPath, "user.email", "e2e@example.com")
	if err := gitSeedCmd2.Run(); err != nil {
		t.Fatalf("failed to seed user.email in temp git config: %v", err)
	}

	// Construct helper function to run the compiled setupper binary
	runSetupper := func(args ...string) (string, string, error) {
		cmd := exec.Command(binaryPath, args...)
		cmd.Env = append(os.Environ(),
			"HOME="+tempDir,
			"GIT_CONFIG_GLOBAL="+gitConfigPath,
		)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Run()
		return stdout.String(), stderr.String(), err
	}

	// Test 1: Run 'init' to scan and create desired.yaml
	t.Run("init command", func(t *testing.T) {
		stdout, stderr, err := runSetupper("init")
		if err != nil {
			t.Fatalf("init command failed: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
		}
		if !strings.Contains(stdout, "Starter desired manifest created") {
			t.Errorf("expected starter manifest confirmation, got stdout: %s", stdout)
		}

		// Verify desired.yaml exists and contains seeded git-config
		desiredPath := filepath.Join(tempDir, ".setupper", "config", "desired.yaml")
		if _, err := os.Stat(desiredPath); os.IsNotExist(err) {
			t.Errorf("expected desired.yaml at %s to exist", desiredPath)
		}
		
		desiredContent, err := os.ReadFile(desiredPath)
		if err != nil {
			t.Fatalf("failed to read desired.yaml: %v", err)
		}
		
		if !strings.Contains(string(desiredContent), "git-config:user.name") {
			t.Errorf("expected desired.yaml to contain git-config:user.name, got: %s", string(desiredContent))
		}
	})

	// Test 2: Run 'scan' command
	t.Run("scan command", func(t *testing.T) {
		stdout, stderr, err := runSetupper("scan")
		if err != nil {
			t.Fatalf("scan command failed: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
		}
		if !strings.Contains(stdout, "Scan complete!") {
			t.Errorf("expected scan complete output, got: %s", stdout)
		}

		// Verify observed.yaml exists
		observedPath := filepath.Join(tempDir, ".setupper", "cache", "observed.yaml")
		if _, err := os.Stat(observedPath); os.IsNotExist(err) {
			t.Errorf("expected observed.yaml at %s to exist", observedPath)
		}
	})

	// Test 3: Run 'inventory' command
	t.Run("inventory command", func(t *testing.T) {
		stdout, stderr, err := runSetupper("inventory")
		if err != nil {
			t.Fatalf("inventory command failed: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
		}
		if !strings.Contains(stdout, "git-config:user.name") {
			t.Errorf("expected inventory to show git-config:user.name, got: %s", stdout)
		}
	})

	// Test 4: Run 'recommend' command
	t.Run("recommend command", func(t *testing.T) {
		stdout, stderr, err := runSetupper("recommend")
		if err != nil {
			t.Fatalf("recommend command failed: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
		}
		// Should output some recommendations or status
		if len(stdout) == 0 {
			t.Errorf("expected recommend output, got empty")
		}
	})

	// Test 5: Run 'plan' command with --yes
	t.Run("plan command", func(t *testing.T) {
		stdout, stderr, err := runSetupper("plan", "--yes")
		if err != nil {
			t.Fatalf("plan command failed: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
		}
		if !strings.Contains(stdout, "Execution plan saved") {
			t.Errorf("expected plan output, got: %s", stdout)
		}

		// Verify plan.yaml exists
		planPath := filepath.Join(tempDir, ".setupper", "cache", "plan.yaml")
		if _, err := os.Stat(planPath); os.IsNotExist(err) {
			t.Errorf("expected plan.yaml to exist")
		}
	})

	// Test 6: Run 'apply' command
	t.Run("apply command", func(t *testing.T) {
		stdout, stderr, err := runSetupper("apply")
		if err != nil {
			t.Fatalf("apply command failed: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
		}
		if !strings.Contains(stdout, "No steps to execute") && !strings.Contains(stdout, "Apply complete") {
			t.Errorf("expected apply output to show completion or no steps, got: %s", stdout)
		}
	})

	// Test 7: Run 'verify' command
	t.Run("verify command", func(t *testing.T) {
		stdout, stderr, err := runSetupper("verify")
		if err != nil {
			t.Fatalf("verify command failed: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
		}
		if !strings.Contains(stdout, "Verification complete") && !strings.Contains(stdout, "verified") {
			t.Errorf("expected verify output, got: %s", stdout)
		}
	})

	// Test 8: Run 'export' command
	t.Run("export command", func(t *testing.T) {
		outputPath := filepath.Join(tempDir, "bootstrap.sh")
		stdout, stderr, err := runSetupper("export", "-f", "shell", "-o", outputPath)
		if err != nil {
			t.Fatalf("export command failed: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
		}
		
		if _, err := os.Stat(outputPath); os.IsNotExist(err) {
			t.Errorf("expected exported script to exist at %s", outputPath)
		}
		
		content, err := os.ReadFile(outputPath)
		if err != nil {
			t.Fatalf("failed to read exported script: %v", err)
		}
		if !strings.Contains(string(content), "#!/bin/zsh") {
			t.Errorf("expected exported script to contain shebang, got: %s", string(content))
		}
	})
}
