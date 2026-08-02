package verifier

import (
	"context"
	"fmt"
	"os/exec"
	"sort"

	"setupper/internal/auth"
	"setupper/internal/manifest"
	"setupper/internal/scanner"
)

type CheckLevel string

const (
	LevelFast CheckLevel = "Fast"
	LevelDeep CheckLevel = "Deep"
)

type Status string

const (
	StatusPassed  Status = "Passed"
	StatusWarning Status = "Warning"
	StatusFailed  Status = "Failed"
)

type CheckResult struct {
	Resource   manifest.Resource
	Status     Status
	Message    string
	CheckLevel CheckLevel
	Key        string
}

type LookPathFunc func(string) (string, error)

type Verifier struct {
	scanner  *scanner.Scanner
	lookPath LookPathFunc
}

func New(s *scanner.Scanner) *Verifier {
	return &Verifier{
		scanner:  s,
		lookPath: exec.LookPath,
	}
}

func NewWithOptions(s *scanner.Scanner, lp LookPathFunc) *Verifier {
	if lp == nil {
		lp = exec.LookPath
	}
	return &Verifier{
		scanner:  s,
		lookPath: lp,
	}
}

// Verify checks the state of the desired manifest against the system
func (v *Verifier) Verify(ctx context.Context, des *manifest.DesiredManifest, deep bool) ([]CheckResult, *manifest.ObservedManifest, error) {
	var results []CheckResult

	// Fast structural check: scan to see what's actually there
	obs, err := v.scanner.Scan(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to scan for verification: %w", err)
	}

	for _, res := range des.Resources {
		key := manifest.FormatKey(res.Type, res.Name)
		if res.Type == "mas" && res.Options["id"] != "" {
			key = manifest.FormatKey("mas", res.Options["id"])
		}

		level := LevelFast
		if deep {
			level = LevelDeep
		}

		_, exists := obs.Resources[key]
		if !exists {
			results = append(results, CheckResult{
				Resource:   res,
				Status:     StatusFailed,
				Message:    "Not installed",
				CheckLevel: LevelFast, // if not installed, it didn't even pass fast
				Key:        key,
			})
			continue
		}

		if deep && res.Authenticated {
			authenticator := auth.GetAuthenticator(res.Name)
			if authenticator != nil {
				isAuth, account, err := authenticator.Check(ctx)
				if isAuth {
					results = append(results, CheckResult{
						Resource:   res,
						Status:     StatusPassed,
						Message:    fmt.Sprintf("Installed & Authenticated (%s)", account),
						CheckLevel: level,
						Key:        key,
					})
				} else {
					results = append(results, CheckResult{
						Resource:   res,
						Status:     StatusFailed,
						Message:    fmt.Sprintf("Auth check failed: %v", err),
						CheckLevel: level,
						Key:        key,
					})
				}
				continue
			} else {
				results = append(results, CheckResult{
					Resource:   res,
					Status:     StatusWarning,
					Message:    "No deep check available",
					CheckLevel: level,
					Key:        key,
				})
				continue
			}
		}

		if res.Type == "brew" {
			_, err := v.lookPath(res.Name)
			if err != nil {
				results = append(results, CheckResult{
					Resource:   res,
					Status:     StatusWarning,
					Message:    "Installed via brew but binary not found in PATH",
					CheckLevel: level,
					Key:        key,
				})
				continue
			}
		}

		if res.Type == "git-config" {
			obsRes, exists := obs.Resources[key]
			if exists {
				desVal := res.Options["value"]
				obsVal := obsRes.Options["value"]
				if desVal != obsVal {
					results = append(results, CheckResult{
						Resource:   res,
						Status:     StatusWarning,
						Message:    fmt.Sprintf("Value mismatch: expected %q, got %q", desVal, obsVal),
						CheckLevel: level,
						Key:        key,
					})
					continue
				}
			}
		}

		msg := "Installed and healthy"
		if res.Type == "cask" || res.Type == "mas" || res.Type == "vscode-extension" || res.Type == "cursor-extension" || res.Type == "font" || res.Type == "application" || res.Type == "brew-tap" {
			msg = "Installed"
		} else if res.Type == "git-config" {
			msg = "Configured correctly"
		}

		results = append(results, CheckResult{
			Resource:   res,
			Status:     StatusPassed,
			Message:    msg,
			CheckLevel: level,
			Key:        key,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].Resource.Type == results[j].Resource.Type {
			return results[i].Resource.Name < results[j].Resource.Name
		}
		return results[i].Resource.Type < results[j].Resource.Type
	})

	return results, obs, nil
}

func Summarize(results []CheckResult) (passed, warnings, failed int) {
	for _, r := range results {
		switch r.Status {
		case StatusPassed:
			passed++
		case StatusWarning:
			warnings++
		case StatusFailed:
			failed++
		}
	}
	return
}
