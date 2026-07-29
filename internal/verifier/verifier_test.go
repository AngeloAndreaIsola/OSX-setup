package verifier

import (
	"context"
	"fmt"
	"testing"

	"setupper/internal/manifest"
	"setupper/internal/runner"
	"setupper/internal/scanner"
)

func fakeLookPath(found map[string]bool) func(string) (string, error) {
	return func(name string) (string, error) {
		if found[name] {
			return "/usr/local/bin/" + name, nil
		}
		return "", fmt.Errorf("not found: %s", name)
	}
}

func TestVerifyAllPresent(t *testing.T) {
	fr := runner.NewFakeRunner()
	fr.Results = map[string][]byte{
		"brew": []byte("git\ngo\n"),
		"mas":  []byte("497799835 Xcode (14.2)\n"),
	}
	s := scanner.New(fr)
	v := NewWithOptions(s, fakeLookPath(map[string]bool{"git": true, "go": true}))

	des := &manifest.DesiredManifest{
		Resources: map[string]manifest.Resource{
			"brew:git":      {Type: "brew", Name: "git"},
			"brew:go":       {Type: "brew", Name: "go"},
			"mas:497799835": {Type: "mas", Name: "Xcode", Options: map[string]string{"id": "497799835"}},
		},
	}

	results, _, err := v.Verify(context.Background(), des, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	for _, res := range results {
		if res.Status != StatusPassed {
			t.Errorf("expected StatusPassed for %s, got %v: %s", res.Key, res.Status, res.Message)
		}
	}
}

func TestVerifyMissing(t *testing.T) {
	fr := runner.NewFakeRunner()
	fr.Results = map[string][]byte{
		"brew": []byte("git\n"),
	}
	s := scanner.New(fr)
	v := NewWithOptions(s, fakeLookPath(map[string]bool{"git": true}))

	des := &manifest.DesiredManifest{
		Resources: map[string]manifest.Resource{
			"brew:git": {Type: "brew", Name: "git"},
			"brew:go":  {Type: "brew", Name: "go"}, // missing
		},
	}

	results, _, err := v.Verify(context.Background(), des, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var missingResult *CheckResult
	for i, res := range results {
		if res.Key == "brew:go" {
			missingResult = &results[i]
			break
		}
	}

	if missingResult == nil {
		t.Fatal("expected to find result for brew:go")
	}

	if missingResult.Status != StatusFailed {
		t.Errorf("expected StatusFailed, got %v", missingResult.Status)
	}
}

func TestVerifyBrewNoBinary(t *testing.T) {
	fr := runner.NewFakeRunner()
	fr.Results = map[string][]byte{
		"brew": []byte("foo\n"),
	}
	s := scanner.New(fr)
	v := NewWithOptions(s, fakeLookPath(map[string]bool{})) // binary not found

	des := &manifest.DesiredManifest{
		Resources: map[string]manifest.Resource{
			"brew:foo": {Type: "brew", Name: "foo"},
		},
	}

	results, _, err := v.Verify(context.Background(), des, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if results[0].Status != StatusWarning {
		t.Errorf("expected StatusWarning, got %v", results[0].Status)
	}
}

func TestVerifyBrewWithBinary(t *testing.T) {
	fr := runner.NewFakeRunner()
	fr.Results = map[string][]byte{
		"brew": []byte("foo\n"),
	}
	s := scanner.New(fr)
	v := NewWithOptions(s, fakeLookPath(map[string]bool{"foo": true}))

	des := &manifest.DesiredManifest{
		Resources: map[string]manifest.Resource{
			"brew:foo": {Type: "brew", Name: "foo"},
		},
	}

	results, _, err := v.Verify(context.Background(), des, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if results[0].Status != StatusPassed {
		t.Errorf("expected StatusPassed, got %v", results[0].Status)
	}
}

func TestVerifyCaskPresent(t *testing.T) {
	fr := runner.NewFakeRunner()
	fr.Results = map[string][]byte{
		"brew": []byte("alfred\n"),
	}
	s := scanner.New(fr)
	v := NewWithOptions(s, fakeLookPath(map[string]bool{}))

	des := &manifest.DesiredManifest{
		Resources: map[string]manifest.Resource{
			"cask:alfred": {Type: "cask", Name: "alfred"},
		},
	}

	results, _, err := v.Verify(context.Background(), des, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if results[0].Status != StatusPassed {
		t.Errorf("expected StatusPassed for cask, got %v", results[0].Status)
	}
}

func TestVerifyDeepAuthSuccess(t *testing.T) {
	fr := runner.NewFakeRunner()
	fr.Results = map[string][]byte{
		"brew": []byte("gh\n"),
	}
	s := scanner.New(fr)
	v := NewWithOptions(s, fakeLookPath(map[string]bool{"gh": true}))

	des := &manifest.DesiredManifest{
		Resources: map[string]manifest.Resource{
			"brew:gh": {Type: "brew", Name: "gh", Authenticated: false},
		},
	}

	results, _, err := v.Verify(context.Background(), des, true) // deep=true
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results[0].Status != StatusPassed {
		t.Errorf("expected StatusPassed, got %v", results[0].Status)
	}
}

func TestVerifyDeepNoAuthenticator(t *testing.T) {
	fr := runner.NewFakeRunner()
	fr.Results = map[string][]byte{
		"brew": []byte("jq\n"),
	}
	s := scanner.New(fr)
	v := NewWithOptions(s, fakeLookPath(map[string]bool{"jq": true}))

	des := &manifest.DesiredManifest{
		Resources: map[string]manifest.Resource{
			"brew:jq": {Type: "brew", Name: "jq", Authenticated: true},
		},
	}

	results, _, err := v.Verify(context.Background(), des, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if results[0].Status != StatusWarning {
		t.Errorf("expected StatusWarning due to no authenticator, got %v", results[0].Status)
	}
}

func TestSummarize(t *testing.T) {
	results := []CheckResult{
		{Status: StatusPassed},
		{Status: StatusFailed},
		{Status: StatusWarning},
		{Status: StatusPassed},
	}
	passed, warnings, failed := Summarize(results)
	if passed != 2 || warnings != 1 || failed != 1 {
		t.Errorf("expected 2 passed, 1 warning, 1 failed, got %d, %d, %d", passed, warnings, failed)
	}
}

func TestVerifyResultsSorted(t *testing.T) {
	fr := runner.NewFakeRunner()
	fr.Results = map[string][]byte{
		"brew": []byte("z\na\n"),
		"mas":  []byte("123 m\n"),
	}
	s := scanner.New(fr)
	v := NewWithOptions(s, fakeLookPath(map[string]bool{"z": true, "a": true, "m": true}))

	des := &manifest.DesiredManifest{
		Resources: map[string]manifest.Resource{
			"brew:z":  {Type: "brew", Name: "z"},
			"brew:a":  {Type: "brew", Name: "a"},
			"cask:c":  {Type: "cask", Name: "c"},
			"mas:123": {Type: "mas", Name: "m", Options: map[string]string{"id": "123"}},
		},
	}

	results, _, err := v.Verify(context.Background(), des, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 4 {
		t.Fatalf("expected 4 results, got %d", len(results))
	}

	expectedKeys := []string{"brew:a", "brew:z", "cask:c", "mas:123"}
	for i, res := range results {
		if res.Key != expectedKeys[i] {
			t.Errorf("at index %d, expected key %s, got %s", i, expectedKeys[i], res.Key)
		}
	}
}

func TestVerifyEmptyManifest(t *testing.T) {
	fr := runner.NewFakeRunner()
	s := scanner.New(fr)
	v := NewWithOptions(s, fakeLookPath(map[string]bool{}))

	des := &manifest.DesiredManifest{
		Resources: map[string]manifest.Resource{},
	}

	results, _, err := v.Verify(context.Background(), des, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}
