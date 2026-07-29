package auth

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

// SecretBackend is designed to allow future pluggable secrets (e.g. 1Password, Keychain)
type SecretBackend interface {
	GetSecret(key string) (string, error)
	// Implementation deferred to future milestone
}

type Authenticator interface {
	Login(ctx context.Context) error
	Check(ctx context.Context) (bool, string, error)
}

type GitHubAuthenticator struct{}

func (g *GitHubAuthenticator) Login(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "gh", "auth", "login")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (g *GitHubAuthenticator) Check(ctx context.Context) (bool, string, error) {
	cmd := exec.CommandContext(ctx, "gh", "auth", "status")
	out, err := cmd.CombinedOutput()
	if err == nil {
		return true, "GitHub Authenticated", nil
	}
	return false, "", fmt.Errorf("gh auth status failed: %s", string(out))
}

type AWSAuthenticator struct{}

func (a *AWSAuthenticator) Login(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "aws", "sso", "login")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (a *AWSAuthenticator) Check(ctx context.Context) (bool, string, error) {
	cmd := exec.CommandContext(ctx, "aws", "sts", "get-caller-identity")
	err := cmd.Run()
	if err == nil {
		return true, "AWS Authenticated", nil
	}
	return false, "", err
}

func GetAuthenticator(toolName string) Authenticator {
	switch toolName {
	case "gh":
		return &GitHubAuthenticator{}
	case "awscli", "aws":
		return &AWSAuthenticator{}
	default:
		return nil
	}
}
