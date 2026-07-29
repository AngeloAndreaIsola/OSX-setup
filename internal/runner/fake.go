package runner

import "context"

// FakeRunner is a fake Runner for tests
type FakeRunner struct {
	Results map[string][]byte
	Errors  map[string]error
	Calls   []Call
}

type Call struct {
	Name string
	Args []string
}

// NewFakeRunner creates a new FakeRunner
func NewFakeRunner() *FakeRunner {
	return &FakeRunner{
		Results: make(map[string][]byte),
		Errors:  make(map[string]error),
	}
}

// Run records the call and returns fake results
func (r *FakeRunner) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	r.Calls = append(r.Calls, Call{Name: name, Args: args})
	
	if err, ok := r.Errors[name]; ok {
		return nil, err
	}
	if out, ok := r.Results[name]; ok {
		return out, nil
	}
	return []byte{}, nil
}
