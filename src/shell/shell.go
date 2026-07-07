package shell

import (
	"context"
	"fmt"
	"mini-codex/src/core"
	"os/exec"
	"time"
)

type ExecRequest struct {
	Command    string
	Args       []string
	CurrentDir string
	TimeoutMs  int
}

type ExecResult struct {
	Stdout     string
	DurationMs int64
}

func RunExec(ctx context.Context, r ExecRequest) <-chan core.Result[ExecResult] {
	resultCh := make(chan core.Result[ExecResult], 1)

	go func() {
		defer close(resultCh)

		var cancel context.CancelFunc

		if r.TimeoutMs != 0 {
			ctx, cancel = context.WithTimeout(ctx, time.Duration(r.TimeoutMs)*time.Millisecond)
			defer cancel()
		}

		started := time.Now()

		workCh := make(chan core.Result[[]byte], 1)

		go func() {
			defer close(workCh)

			cmd := exec.CommandContext(ctx, r.Command, r.Args...)
			cmd.Dir = r.CurrentDir

			output, err := cmd.Output()
			if err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok {
					err = fmt.Errorf("command failed with exit code %d: %s", exitErr.ExitCode(), string(exitErr.Stderr))
					workCh <- core.Result[[]byte]{Err: err}
					return
				}
				workCh <- core.Result[[]byte]{Err: err}
				return
			}
			workCh <- core.Result[[]byte]{Val: output}
		}()

		select {
		case output := <-workCh:
			if output.Err != nil {
				resultCh <- core.Result[ExecResult]{Err: output.Err}
				return
			}
			result := ExecResult{
				Stdout:     string(output.Val),
				DurationMs: time.Since(started).Milliseconds(),
			}
			resultCh <- core.Result[ExecResult]{Val: result}
		case <-ctx.Done():
			resultCh <- core.Result[ExecResult]{Err: ctx.Err()}
		}
	}()

	return resultCh
}
