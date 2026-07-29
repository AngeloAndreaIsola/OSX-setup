package runner_test

import (
	"context"
	"testing"

	"setupper/internal/runner"
)

func TestSubprocessRunner(t *testing.T) {
	r := runner.NewSubprocessRunner()
	out, err := r.Run(context.Background(), "echo", "hello")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if string(out) != "hello\n" {
		t.Errorf("expected 'hello\\n', got %q", string(out))
	}
}

func TestFakeRunner(t *testing.T) {
	r := runner.NewFakeRunner()
	r.Results["echo"] = []byte("fake hello\n")
	
	out, err := r.Run(context.Background(), "echo", "hello")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if string(out) != "fake hello\n" {
		t.Errorf("expected 'fake hello\\n', got %q", string(out))
	}
	
	if len(r.Calls) != 1 {
		t.Errorf("expected 1 call, got %d", len(r.Calls))
	}
	if r.Calls[0].Name != "echo" || len(r.Calls[0].Args) != 1 || r.Calls[0].Args[0] != "hello" {
		t.Errorf("unexpected call: %+v", r.Calls[0])
	}
}
