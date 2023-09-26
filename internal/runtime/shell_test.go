package runtime

import (
	"context"
	"testing"
	"time"

	"github.com/runabol/tork"
	"github.com/runabol/tork/internal/uuid"
	"github.com/stretchr/testify/assert"
)

func TestShellRuntimeRunResult(t *testing.T) {
	rt := NewShellRuntime()

	tk := &tork.Task{
		ID:  uuid.NewUUID(),
		Run: "echo -n hello world > $TORK_OUTPUT",
	}

	err := rt.Run(context.Background(), tk)

	assert.NoError(t, err)
	assert.Equal(t, "hello world", tk.Result)
}

func TestShellRuntimeCmd(t *testing.T) {
	rt := NewShellRuntime()

	tk := &tork.Task{
		ID:  uuid.NewUUID(),
		CMD: []string{"echo", "hello", "world"},
	}

	err := rt.Run(context.Background(), tk)

	assert.NoError(t, err)
}

func TestShellRuntimeRunFile(t *testing.T) {
	rt := NewShellRuntime()

	tk := &tork.Task{
		ID:  uuid.NewUUID(),
		Run: "cat hello.txt > $TORK_OUTPUT",
		Files: map[string]string{
			"hello.txt": "hello world",
		},
	}

	err := rt.Run(context.Background(), tk)

	assert.NoError(t, err)
	assert.Equal(t, "hello world", tk.Result)
}

func TestShellRuntimeRunNotSupported(t *testing.T) {
	rt := NewShellRuntime()

	tk := &tork.Task{
		ID:       uuid.NewUUID(),
		Run:      "echo hello world",
		Networks: []string{"some-network"},
	}

	err := rt.Run(context.Background(), tk)

	assert.Error(t, err)
}

func TestShellRuntimeRunError(t *testing.T) {
	rt := NewShellRuntime()

	tk := &tork.Task{
		ID:  uuid.NewUUID(),
		Run: "no_such_command",
	}

	err := rt.Run(context.Background(), tk)

	assert.Error(t, err)
}

func TestShellRuntimeRunTimeout(t *testing.T) {
	rt := NewShellRuntime()

	tk := &tork.Task{
		ID:  uuid.NewUUID(),
		Run: "sleep 30",
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()

	err := rt.Run(ctx, tk)

	assert.Error(t, err)
}

func TestShellRuntimeStop(t *testing.T) {
	rt := NewShellRuntime()

	tk := &tork.Task{
		ID:  uuid.NewUUID(),
		Run: "sleep 5",
	}

	ch := make(chan any)

	go func() {
		err := rt.Run(context.Background(), tk)
		assert.Error(t, err)
		close(ch)
	}()

	time.Sleep(time.Second * 1)

	err := rt.Stop(context.Background(), tk)
	assert.NoError(t, err)
	<-ch
}
