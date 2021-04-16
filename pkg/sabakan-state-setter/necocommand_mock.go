package sss

import (
	"context"
)

type necoCmdMockExecutor struct {
	tpm map[string]int
}

func newMockNecoCmdExecutor() *necoCmdMockExecutor {
	return &necoCmdMockExecutor{
		tpm: map[string]int{},
	}
}

func (e *necoCmdMockExecutor) TPMClear(ctx context.Context, serial string) ([]byte, error) {
	e.tpm[serial]++
	return []byte("log message"), nil
}

// test function
func (e *necoCmdMockExecutor) getTPMClearCount(serial string) int {
	return e.tpm[serial]
}
