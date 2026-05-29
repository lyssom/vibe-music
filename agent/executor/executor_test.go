package executor_test

import (
	"context"
	"testing"

	"github.com/lyssom/vibe-music/agent/executor"
)

type mockExecutor struct{}

func (m *mockExecutor) Execute(_ context.Context, code string) executor.ExecResult {
	if code == "bad" {
		return executor.ExecResult{
			Success: false,
			Error:   "syntax error",
		}
	}
	return executor.ExecResult{
		Success: true,
		Output:  "playing...",
	}
}

func (m *mockExecutor) Stop() {}

var _ executor.Executor = (*mockExecutor)(nil)

func TestExecutorSuccess(t *testing.T) {
	e := &mockExecutor{}
	result := e.Execute(context.Background(), `sound("bd")`)
	if !result.Success {
		t.Error("expected success")
	}
	if result.Output != "playing..." {
		t.Errorf("expected 'playing...', got %q", result.Output)
	}
}

func TestExecutorFailure(t *testing.T) {
	e := &mockExecutor{}
	result := e.Execute(context.Background(), "bad")
	if result.Success {
		t.Error("expected failure")
	}
	if result.Error != "syntax error" {
		t.Errorf("expected 'syntax error', got %q", result.Error)
	}
}

func TestExecutorStop(t *testing.T) {
	e := &mockExecutor{}
	e.Stop() // should not panic
}