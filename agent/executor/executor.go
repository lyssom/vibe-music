// Package executor defines the interface for executing pattern code.
package executor

import "context"

// ExecResult holds the outcome of a code execution attempt.
type ExecResult struct {
	// Success indicates whether the code ran without errors.
	Success bool
	// Error contains the error message if execution failed.
	Error string
	// Output contains any stdout output from execution.
	Output string
}

// Executor runs pattern code against the audio engine.
type Executor interface {
	// Execute runs the given code and returns the result.
	Execute(ctx context.Context, code string) ExecResult

	// Stop halts any currently running execution.
	Stop()
}